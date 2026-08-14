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
//     base64 image uploads, chunked so a chunk always fits the MP GC heap
//   - ferret_image_manifest / ferret_delete_image: the image manifest
//   - ferret_draw_image: reads the file (or the sprite cache), blits with
//     chroma keying, using the manifest for the dimensions
//
// Source files are capped at maxFsFile so a file always fits the 16 KiB MP GC
// heap once compiled; the image cache is bounded to a fixed RAM budget because
// the display framebuffer already owns 115 KiB of the 264 KiB.

package main

/*
#include <stdint.h>
*/
import "C"

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"os"
	"sort"
	"strings"
	"unsafe"

	"tinygo.org/x/tinyfs/littlefs"

	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
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
	// maxImageFile caps a single image (raw RGB565 bytes).
	maxImageFile = 64 * 1024

	imgDirPath       = "/img"
	imgManifestPath  = "/img/images.txt"
	imgCacheBudget   = 48 * 1024 // RAM budget for decoded sprites
)

var (
	ferretFs *littlefs.LFS

	// imgManifest: image name -> dims + FNV-1a checksum of the uploaded
	// base64 string (the web editor's diff format). Persisted to
	// /img/images.txt so dims survive a reboot.
	imgManifest = map[string]imgMeta{}

	// sprite cache: decoded RGB565 blobs, FIFO-evicted against a byte budget.
	imgCache = spriteCache{entries: map[string]cachedImg{}}

	// editor uploads arrive as base64 chunks; accumulate until
	// ferret_write_image_end validates and persists.
	pendingImgs = map[string]*pendingImage{}
)

type imgMeta struct {
	w, h   int
	chksum string
}

type cachedImg struct {
	data  []byte
	w, h  uint8
	order int
}

type spriteCache struct {
	entries map[string]cachedImg
	order   []string // insertion order, oldest first
	bytes   int
	next    int
}

type pendingImage struct {
	w, h int
	b64  []byte // base64 chars accumulate here; decoded once at write_image_end
	f    hash.Hash32 // FNV-1a over the base64 chunks, in order
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

//export ferret_stat
// Returns MP_IMPORT_STAT_*: 0 = missing, 1 = dir, 2 = file, -1 = error.
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

//export ferret_read_file
// Returns the content length, 0 if missing, -1 on error, -2 if too big.
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

//export ferret_write_file
// Replaces the file. 0 = ok, -1 = error, -2 = too big.
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

//export ferret_append_file
// Appends to the file (creating it if needed). Same return codes as write.
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
	return name != "" && !strings.ContainsAny(name, "/\\:") && name != "." && name != ".."
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
		imgManifest[fields[0]] = imgMeta{w: w, h: h, chksum: fields[3]}
	}
}

//export ferret_write_image
// Starts (or restarts) a base64 upload. 0 = ok, -1 = invalid args.
func ferret_write_image(name *C.char, w, h C.int, b64 *C.char, n C.int) C.int {
	nm := C.GoString(name)
	if !validImageName(nm) || int(w) <= 0 || int(h) <= 0 || int(w)*int(h)*2 > maxImageFile {
		return -1
	}
	p := &pendingImage{w: int(w), h: int(h), f: fnv.New32a()}
	if !pendingAppend(p, unsafe.Slice((*byte)(unsafe.Pointer(b64)), int(n))) {
		return -1
	}
	pendingImgs[nm] = p
	return 0
}

//export ferret_append_image
// Appends a base64 chunk to an in-progress upload.
// 0 = ok, -1 = invalid, -2 = no upload started, -3 = exceeds maxImageFile.
func ferret_append_image(name *C.char, b64 *C.char, n C.int) C.int {
	nm := C.GoString(name)
	p, ok := pendingImgs[nm]
	if !ok {
		return -2
	}
	if (len(p.b64)+int(n))*3/4 > maxImageFile {
		delete(pendingImgs, nm)
		return -3
	}
	if !pendingAppend(p, unsafe.Slice((*byte)(unsafe.Pointer(b64)), int(n))) {
		delete(pendingImgs, nm)
		return -1
	}
	return 0
}

// pendingAppend accumulates base64 characters (the editor splits the payload
// at arbitrary chunk boundaries, which are not valid base64 on their own) and
// feeds them to the checksum hash. Decoding happens once at write_image_end,
// when the full, padded string is available.
func pendingAppend(p *pendingImage, chunk []byte) bool {
	p.b64 = append(p.b64, chunk...)
	p.f.Write(chunk)
	return true
}

//export ferret_write_image_end
// Validates and persists a finished upload. 0 = ok, -1 = bad base64/size, -2 =
// no upload started.
func ferret_write_image_end(name *C.char) C.int {
	nm := C.GoString(name)
	p, ok := pendingImgs[nm]
	if !ok {
		return -2
	}
	delete(pendingImgs, nm)
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(p.b64)))
	m, err := base64.StdEncoding.Decode(decoded, p.b64)
	if err != nil || m != p.w*p.h*2 {
		return -1
	}
	if err := fsWrite(imgDirPath+"/"+nm, decoded[:m]); err != nil {
		return -1
	}
	imgManifest[nm] = imgMeta{
		w:      p.w,
		h:      p.h,
		chksum: fmt.Sprintf("%08x", p.f.Sum32()),
	}
	imgCache.evict(nm)
	saveManifest()
	return 0
}

//export ferret_image_manifest
// Returns the manifest as "name,width,height,checksum" lines. Returns the
// length, or -1 if the buffer is too small.
func ferret_image_manifest(buf *C.char, max C.int) C.int {
	s := manifestString()
	if len(s) > int(max) {
		return -1
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(max)), s)
	return C.int(len(s))
}

//export ferret_delete_image
// Removes an image and its manifest entry. 0 = ok, -1 = invalid name/error.
func ferret_delete_image(name *C.char) C.int {
	nm := C.GoString(name)
	if !validImageName(nm) {
		return -1
	}
	if err := ferretFs.Remove(absPath(imgDirPath + "/" + nm)); err != nil && !isMissing(err) {
		return -1
	}
	delete(imgManifest, nm)
	imgCache.evict(nm)
	saveManifest()
	return 0
}

// imageData returns the decoded RGB565 bytes for a manifest entry, drawing
// from the sprite cache when it fits the budget.
func imageData(name string, m imgMeta) ([]byte, error) {
	if c, ok := imgCache.entries[name]; ok {
		return c.data, nil
	}
	data, err := fsRead(imgDirPath + "/" + name)
	if err != nil {
		return nil, err
	}
	if len(data) != m.w*m.h*2 {
		return nil, errors.New("image file size does not match manifest")
	}
	if len(data) <= imgCacheBudget {
		imgCache.insert(name, data, uint8(m.w), uint8(m.h))
	}
	return data, nil
}

func (c *spriteCache) insert(name string, data []byte, w, h uint8) {
	for c.bytes+len(data) > imgCacheBudget && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		if e, ok := c.entries[oldest]; ok {
			c.bytes -= len(e.data)
			delete(c.entries, oldest)
		}
	}
	if _, ok := c.entries[name]; ok {
		c.bytes -= len(c.entries[name].data)
	}
	c.entries[name] = cachedImg{data: data, w: w, h: h, order: c.next}
	c.order = append(c.order, name)
	c.bytes += len(data)
	c.next++
}

func (c *spriteCache) evict(name string) {
	if e, ok := c.entries[name]; ok {
		c.bytes -= len(e.data)
		delete(c.entries, name)
	}
	for i, n := range c.order {
		if n == name {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

//export ferret_draw_image
// Blits an image by manifest name. 0 = ok, -1 = unknown image, -2 = too big,
// -3 = read error.
func ferret_draw_image(name *C.char, x, y C.int) C.int {
	nm := C.GoString(name)
	m, ok := imgManifest[nm]
	if !ok {
		return -1
	}
	if m.w > 255 || m.h > 255 {
		return -2
	}
	data, err := imageData(nm, m)
	if err != nil {
		return -3
	}
	helpers.DrawImage(display, helpers.Image{
		Data: string(data),
		W:    uint8(m.w),
		H:    uint8(m.h),
	}, int16(x), int16(y))
	return 0
}