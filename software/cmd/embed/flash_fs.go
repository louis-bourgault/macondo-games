// flash_fs.go — the Go-owned LittleFS2 volume and image service.
//
// MicroPython never touches the flash filesystem directly; the volume is
// mounted here with tinygo.org/x/tinyfs/littlefs (the same LittleFS2 format
// the old port/flash_storage.c used, so existing volumes keep working). The C
// layer reaches it only through the ferret_* exports:
//
//   - ferret_get_boot_script: /main.py at boot; host.c compiles and runs it
//     right after mp_embed_init, so a soft reset (Ctrl-D) reboots the game
//   - ferret_stat / ferret_read_file: module imports (mp_import_stat and
//     mp_lexer_new_from_file in port/repl_stubs.c)
//   - ferret_write_file / ferret_append_file: the web editor's code files
//   - ferret_write_image / ferret_append_image / ferret_write_image_end:
//     base64 image uploads, decoded incrementally into a temporary file
//   - ferret_image_manifest / ferret_delete_image: the image manifest
//   - ferret_draw_image: streams the file a row at a time and blits with chroma
//     keying, using the manifest for the dimensions
//
// Source files are capped at maxFsFile so a file always fits the 16 KiB MP GC
// heap once compiled. Filesystem images are deliberately streamed because the
// display framebuffer already owns 115 KiB of the RP2040's RAM.

package main

/*
#include <stdint.h>
*/
import "C"

import (
	"encoding/base64"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"os"
	"sort"
	"strings"
	"unsafe"

	"tinygo.org/x/tinyfs"
	"tinygo.org/x/tinyfs/littlefs"
)

const (
	// lfs tuning: 256 B cache buffers (read+prog), lookahead covers the whole
	// 2048-block volume (2048 bits = 256 B).
	fsCacheSize     = 256
	fsLookaheadSize = 256
	fsBlockCycles   = 500

	// maxFsFile caps source files (boot script, modules). A file this big
	// already fills the MP GC heap once compiled, so nothing useful is lost.
	maxFsFile = 16 * 1024
	// Images are stored as little-endian RGB565 and may cover the whole display.
	// Validate the dimensions before multiplying: int is 32-bit on the device.
	maxImageWidth  = 240
	maxImageHeight = 240
	maxImageFile   = maxImageWidth * maxImageHeight * 2

	imgDirPath        = "/img"
	imgManifestPath   = "/img/images.txt"
	imgUploadTempName = ".upload.tmp"
	imgUploadTempPath = imgDirPath + "/" + imgUploadTempName

	// The editor currently sends 320 base64 characters per call, which decode to
	// 240 bytes. Process larger manual calls in equally bounded pieces.
	imageDecodeBlock = 320
)

var (
	ferretFs *littlefs.LFS

	// imgManifest: image name -> dims + FNV-1a checksum of the uploaded
	// base64 string (the web editor's diff format). Persisted to
	// /img/images.txt so dims survive a reboot.
	imgManifest = map[string]imgMeta{}

	// The editor uploads one image at a time. Keeping a single open temporary
	// file bounds both RAM and LittleFS file-cache use.
	pendingImg *pendingImage

	// Shared by upload decoding (first 240 bytes) and row-based drawing (all 480
	// bytes). MicroPython invokes image functions synchronously, so they cannot
	// use this buffer concurrently.
	imageScratch [maxImageWidth * 2]byte
)

type imgMeta struct {
	w, h   int
	chksum string
}

type pendingImage struct {
	name            string
	w, h            int
	expectedDecoded int
	expectedEncoded int
	encoded         int
	decoded         int
	carry           [4]byte
	carryLen        int
	f               hash.Hash32 // FNV-1a over the base64 chunks, in order
	file            tinyfs.File
}

// --- volume setup -----------------------------------------------------------

// initGoFS mounts the LittleFS2 volume over the flash region, formatting it on
// first boot (or when corrupt), then loads the persisted image manifest.
func initGoFS() {
	ferretFs = littlefs.New(flashBlockDev{}).Configure(&littlefs.Config{
		CacheSize:     fsCacheSize,
		LookaheadSize: fsLookaheadSize,
		BlockCycles:   fsBlockCycles,
	})
	if !fsSuperblockPresent() {
		if err := ferretFs.Format(); err != nil {
			panic("ferret: lfs format: " + err.Error())
		}
	}
	if err := ferretFs.Mount(); err != nil {
		panic("ferret: lfs mount: " + err.Error())
	}
	if err := ferretFs.Mkdir(imgDirPath, 0o777); err != nil && !isMissing(err) && !isExists(err) {
		panic("ferret: lfs mkdir /img: " + err.Error())
	}
	// An upload is only published by rename at write_image_end. A reset before
	// that point can leave an unreferenced temporary file behind.
	if err := ferretFs.Remove(imgUploadTempPath); err != nil && !isMissing(err) {
		fmt.Printf("ferret: remove stale image upload: %v\n", err)
	}
	loadManifest()
}

// fsSuperblockPresent detects an existing LittleFS2 volume: the superblock
// magic "littlefs" sits at byte 8 of block 0, between 0xFF bytes.
func fsSuperblockPresent() bool {
	var head [16]byte
	if _, err := (flashBlockDev{}).ReadAt(head[:], 0); err != nil {
		return false
	}
	return string(head[8:16]) == "littlefs"
}

// --- low-level file helpers -------------------------------------------------

// absPath maps a cleaned relative path onto the volume ("main.py" -> "/main.py").
func absPath(p string) string {
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

// cleanPath normalises a path coming from MicroPython/editor: strips the
// leading "/", drops "." and empty segments, and rejects anything that could
// escape the volume ("..", backslashes, colons). Empty result = invalid.
func cleanPath(p string) string {
	parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." || strings.ContainsAny(part, "\\:") {
			return ""
		}
		out = append(out, part)
	}
	return strings.Join(out, "/")
}

// LFS_ERR_NOENT (-2) and LFS_ERR_EXIST (-17): the only tinyfs errors the Go
// layer treats specially.
var (
	errNoEntry = littlefs.Error(-2)
	errExists  = littlefs.Error(-17)
)

func isMissing(err error) bool {
	return err == errNoEntry
}

func isExists(err error) bool {
	return err == errExists
}

func fsWrite(p string, data []byte) error {
	f, err := ferretFs.OpenFile(absPath(p), os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		return err
	}
	defer f.Close()
	// tinyfs's lfs wrapper indexes buf[0] even for empty writes; skip them.
	if len(data) == 0 {
		return nil
	}
	_, err = f.Write(data)
	return err
}

func fsAppend(p string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	f, err := ferretFs.OpenFile(absPath(p), os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

func fsRead(p string) ([]byte, error) {
	info, err := ferretFs.Stat(absPath(p))
	if err != nil {
		return nil, err
	}
	data := make([]byte, info.Size())
	if info.Size() == 0 {
		return data, nil
	}
	f, err := ferretFs.Open(absPath(p))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Read(data)
	return data, err
}

// --- boot script ------------------------------------------------------------

//export ferret_get_boot_script
func ferret_get_boot_script(buf *C.char, max C.int) C.int {
	data, err := fsRead("/main.py")
	if err != nil {
		if isMissing(err) {
			return 0 // no boot script: drop straight into the REPL
		}
		return -2
	}
	if len(data) > int(max) {
		return -1 // bigger than the MP heap can compile anyway
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(max)), data)
	return C.int(len(data))
}

// --- module imports (mp_import_stat / mp_lexer_new_from_file) ----------------

// Returns MP_IMPORT_STAT_*: 0 = missing, 1 = dir, 2 = file, -1 = error.
//
//export ferret_stat
func ferret_stat(path *C.char) C.int {
	p := cleanPath(C.GoString(path))
	if p == "" {
		return 0
	}
	info, err := ferretFs.Stat(absPath(p))
	if err != nil {
		return 0
	}
	if info.IsDir() {
		return 1
	}
	return 2
}

// Returns the content length, 0 if missing, -1 on error, -2 if too big.
//
//export ferret_read_file
func ferret_read_file(path *C.char, buf *C.char, max C.int) C.int {
	p := cleanPath(C.GoString(path))
	if p == "" {
		return -1
	}
	data, err := fsRead(p)
	if err != nil {
		if isMissing(err) {
			return 0
		}
		return -1
	}
	if len(data) > int(max) {
		return -2
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(max)), data)
	return C.int(len(data))
}

// --- code files (web editor) ------------------------------------------------

// Replaces the file. 0 = ok, -1 = error, -2 = too big.
//
//export ferret_write_file
func ferret_write_file(path *C.char, data *C.char, n C.int) C.int {
	p := cleanPath(C.GoString(path))
	if p == "" || n > maxFsFile {
		return -1
	}
	if err := fsWrite(p, unsafe.Slice((*byte)(unsafe.Pointer(data)), int(n))); err != nil {
		return -1
	}
	return 0
}

// Appends to the file (creating it if needed). Same return codes as write.
//
//export ferret_append_file
func ferret_append_file(path *C.char, data *C.char, n C.int) C.int {
	p := cleanPath(C.GoString(path))
	if p == "" {
		return -1
	}
	if info, err := ferretFs.Stat(absPath(p)); err == nil {
		if info.Size()+int64(n) > maxFsFile {
			return -2
		}
	} else if !isMissing(err) {
		return -1
	}
	if err := fsAppend(p, unsafe.Slice((*byte)(unsafe.Pointer(data)), int(n))); err != nil {
		return -1
	}
	return 0
}

// --- image service ----------------------------------------------------------

func validImageName(name string) bool {
	return name != "" && name != imgUploadTempName &&
		!strings.ContainsAny(name, "/\\:,\r\n") && name != "." && name != ".."
}

func manifestString() string {
	names := make([]string, 0, len(imgManifest))
	for name := range imgManifest {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		m := imgManifest[name]
		fmt.Fprintf(&b, "%s,%d,%d,%s\n", name, m.w, m.h, m.chksum)
	}
	return b.String()
}

func saveManifest() {
	if err := fsWrite(imgManifestPath, []byte(manifestString())); err != nil {
		fmt.Printf("ferret: save manifest: %v\n", err)
	}
}

func loadManifest() {
	data, err := fsRead(imgManifestPath)
	if err != nil {
		return // no manifest yet
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, ",")
		if len(fields) != 4 {
			continue
		}
		var w, h int
		if _, err := fmt.Sscanf(fields[1], "%d", &w); err != nil {
			continue
		}
		if _, err := fmt.Sscanf(fields[2], "%d", &h); err != nil {
			continue
		}
		if !validImageName(fields[0]) || w <= 0 || h <= 0 || w > maxImageWidth || h > maxImageHeight {
			continue
		}
		imgManifest[fields[0]] = imgMeta{w: w, h: h, chksum: fields[3]}
	}
}

// abortPendingImage closes and removes an unpublished upload. The previously
// published image is never opened with O_TRUNC, so it remains usable.
func abortPendingImage() {
	if pendingImg == nil {
		return
	}
	if pendingImg.file != nil {
		_ = pendingImg.file.Close()
	}
	pendingImg = nil
	if err := ferretFs.Remove(imgUploadTempPath); err != nil && !isMissing(err) {
		fmt.Printf("ferret: remove failed image upload: %v\n", err)
	}
}

// writeAll handles short writes without allocating. A nil error with no
// progress is treated as an I/O failure rather than spinning forever.
func writeAll(f tinyfs.File, data []byte) bool {
	for len(data) > 0 {
		n, err := f.Write(data)
		if err != nil || n <= 0 || n > len(data) {
			return false
		}
		data = data[n:]
	}
	return true
}

// decodeImageBlock decodes complete base64 quartets and appends the resulting
// RGB565 bytes to the open temporary file. Returns -1 for malformed data, -3
// for a size overrun, and -4 for an I/O failure.
func decodeImageBlock(p *pendingImage, encoded []byte) int {
	n, err := base64.StdEncoding.Decode(imageScratch[:], encoded)
	if err != nil {
		return -1
	}
	if p.decoded+n > p.expectedDecoded {
		return -3
	}
	if !writeAll(p.file, imageScratch[:n]) {
		return -4
	}
	p.decoded += n
	return 0
}

// pendingAppend hashes and incrementally decodes one editor chunk. Only the
// incomplete final base64 quartet is retained between calls.
func pendingAppend(p *pendingImage, chunk []byte) int {
	if p.encoded+len(chunk) > p.expectedEncoded {
		return -3
	}
	_, _ = p.f.Write(chunk)
	p.encoded += len(chunk)

	if p.carryLen > 0 {
		n := 4 - p.carryLen
		if n > len(chunk) {
			n = len(chunk)
		}
		copy(p.carry[p.carryLen:], chunk[:n])
		p.carryLen += n
		chunk = chunk[n:]
		if p.carryLen == 4 {
			if r := decodeImageBlock(p, p.carry[:]); r != 0 {
				return r
			}
			p.carryLen = 0
		}
	}

	for len(chunk) >= 4 {
		n := len(chunk) &^ 3
		if n > imageDecodeBlock {
			n = imageDecodeBlock
		}
		if r := decodeImageBlock(p, chunk[:n]); r != 0 {
			return r
		}
		chunk = chunk[n:]
	}

	copy(p.carry[:], chunk)
	p.carryLen = len(chunk)
	return 0
}

// Starts (or restarts) a base64 upload. 0 = ok, -1 = invalid args, -3 = too
// large, -4 = I/O error.
//
//export ferret_write_image
func ferret_write_image(name *C.char, w, h C.int, b64 *C.char, n C.int) C.int {
	nm := C.GoString(name)
	wv, hv := int(w), int(h)
	if !validImageName(nm) || wv <= 0 || hv <= 0 || int(n) < 0 {
		return -1
	}
	if wv > maxImageWidth || hv > maxImageHeight {
		return -3
	}
	expected := wv * hv * 2
	if expected > maxImageFile {
		return -3
	}
	abortPendingImage()
	f, err := ferretFs.OpenFile(imgUploadTempPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		return -4
	}
	pendingImg = &pendingImage{
		name:            nm,
		w:               wv,
		h:               hv,
		expectedDecoded: expected,
		expectedEncoded: base64.StdEncoding.EncodedLen(expected),
		f:               fnv.New32a(),
		file:            f,
	}
	if r := pendingAppend(pendingImg, unsafe.Slice((*byte)(unsafe.Pointer(b64)), int(n))); r != 0 {
		abortPendingImage()
		return C.int(r)
	}
	return 0
}

// Appends a base64 chunk to an in-progress upload.
// 0 = ok, -1 = invalid base64, -2 = no matching upload, -3 = too large, -4 =
// I/O error.
//
//export ferret_append_image
func ferret_append_image(name *C.char, b64 *C.char, n C.int) C.int {
	nm := C.GoString(name)
	if pendingImg == nil || pendingImg.name != nm {
		return -2
	}
	if int(n) < 0 {
		abortPendingImage()
		return -1
	}
	if r := pendingAppend(pendingImg, unsafe.Slice((*byte)(unsafe.Pointer(b64)), int(n))); r != 0 {
		abortPendingImage()
		return C.int(r)
	}
	return 0
}

// Validates and publishes a finished upload. 0 = ok, -1 = bad base64/size, -2
// = no matching upload, -4 = I/O error.
//
//export ferret_write_image_end
func ferret_write_image_end(name *C.char) C.int {
	nm := C.GoString(name)
	if pendingImg == nil || pendingImg.name != nm {
		return -2
	}
	p := pendingImg
	if p.encoded != p.expectedEncoded || p.carryLen != 0 || p.decoded != p.expectedDecoded {
		abortPendingImage()
		return -1
	}
	if err := p.file.Close(); err != nil {
		p.file = nil
		abortPendingImage()
		return -4
	}
	p.file = nil
	pendingImg = nil
	info, err := ferretFs.Stat(imgUploadTempPath)
	if err != nil || info.Size() != int64(p.expectedDecoded) {
		_ = ferretFs.Remove(imgUploadTempPath)
		return -4
	}
	if err := ferretFs.Rename(imgUploadTempPath, absPath(imgDirPath+"/"+nm)); err != nil {
		_ = ferretFs.Remove(imgUploadTempPath)
		return -4
	}
	imgManifest[nm] = imgMeta{
		w:      p.w,
		h:      p.h,
		chksum: fmt.Sprintf("%08x", p.f.Sum32()),
	}
	saveManifest()
	return 0
}

// Returns the manifest as "name,width,height,checksum" lines. Returns the
// length, or -1 if the buffer is too small.
//
//export ferret_image_manifest
func ferret_image_manifest(buf *C.char, max C.int) C.int {
	s := manifestString()
	if len(s) > int(max) {
		return -1
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(max)), s)
	return C.int(len(s))
}

// Removes an image and its manifest entry. 0 = ok, -1 = invalid name/error.
//
//export ferret_delete_image
func ferret_delete_image(name *C.char) C.int {
	nm := C.GoString(name)
	if !validImageName(nm) {
		return -1
	}
	if err := ferretFs.Remove(absPath(imgDirPath + "/" + nm)); err != nil && !isMissing(err) {
		return -1
	}
	delete(imgManifest, nm)
	saveManifest()
	return 0
}

// Blits an image by manifest name. 0 = ok, -1 = unknown image, -2 = too big,
// -3 = read error.
//
//export ferret_draw_image
func ferret_draw_image(name *C.char, x, y C.int) C.int {
	nm := C.GoString(name)
	m, ok := imgManifest[nm]
	if !ok {
		return -1
	}
	if m.w <= 0 || m.h <= 0 || m.w > maxImageWidth || m.h > maxImageHeight {
		return -2
	}
	info, err := ferretFs.Stat(absPath(imgDirPath + "/" + nm))
	if err != nil || info.Size() != int64(m.w*m.h*2) {
		return -3
	}
	f, err := ferretFs.Open(absPath(imgDirPath + "/" + nm))
	if err != nil {
		return -3
	}
	defer f.Close()

	rowBytes := imageScratch[:m.w*2]
	for row := 0; row < m.h; row++ {
		if _, err := io.ReadFull(f, rowBytes); err != nil {
			return -3
		}
		dy := int(y) + row
		if dy < 0 || dy >= maxImageHeight {
			continue
		}
		for col := 0; col < m.w; col++ {
			dx := int(x) + col
			if dx < 0 || dx >= maxImageWidth {
				continue
			}
			i := col * 2
			color := uint16(rowBytes[i]) | uint16(rowBytes[i+1])<<8
			display.Pixel(dx, dy, color)
		}
	}
	return 0
}
