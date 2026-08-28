// flash_common.go — the Ferretboard flash block device.
//
// layout: the firmware programs live in the first part of the 16 MiB W25Q128
// QSPI flash; the last 8 MiB are reserved for the user filesystem ("assets and
// user code"). That region is presented to the Go filesystem layer
// (flash_fs.go) as a block device with blocks of 4096 bytes (the flash
// erase-sector size), mounted with tinygo.org/x/tinyfs/littlefs.
//
// The raw I/O is per-build: on the device it is machine.Flash (TinyGo's
// driver for the RP2040/RP2350 bootrom flash routines); on the host reference
// it is a plain-file image (flash_stub.go).

package main

const (
	// flashBlockSize matches the W25Q erase sector (4 KiB). Bigger beats
	// smaller here: littlefs and the editor both then work in whole sectors,
	// so writes never pay read-modify-write at the filesystem boundary.
	flashBlockSize = 4096
	// flashStorageBytes is the size of the user filesystem region: the last
	// 8 MiB of the 16 MiB chip.
	flashStorageBytes = 8 * 1024 * 1024
	// writePage is the flash page-program granularity; writes must be
	// page-aligned and page-multiples.
	writePage = 256

	flashRegionBlocks = flashStorageBytes / flashBlockSize
)

// flashBlockDev maps the FS region to a zero-based block device, hiding the
// region's offset from the littlefs mount in flash_fs.go.
type flashBlockDev struct{}

// rawFlash does the actual device I/O; provided per build (machine.Flash on
// device, a file image on the host).
type rawFlash interface {
	ReadAt(p []byte, off int64) (int, error)
	WriteAt(p []byte, off int64) (int, error)
	EraseBlocks(start, length int64) error
	Size() int64
}

var (
	rawBackend   rawFlash
	regionBase   int64 // offset of the FS region within the raw device
	rawTotalSize int64
)

func (d flashBlockDev) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off+int64(len(p)) > flashStorageBytes {
		return 0, flashErrOutOfRange
	}
	return rawBackend.ReadAt(p, regionBase+off)
}

// WriteAt pads partial writes to whole 256-byte pages with 0xFF. NOR flash
// programming only ever clears bits, and littlefs writes only into erased
// sectors, so the padding is always a no-op on surrounding data and no
// read-modify-write is needed.
func (d flashBlockDev) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 || off+int64(len(p)) > flashStorageBytes {
		return 0, flashErrOutOfRange
	}
	if len(p) == 0 {
		return 0, nil
	}
	devOff := regionBase + off
	pageStart := (devOff / writePage) * writePage
	prefix := int(devOff - pageStart)
	total := ((prefix + len(p)) + writePage - 1) / writePage * writePage
	if total > len(padBuf) {
		panic("ferret: flash write spans more than two pages")
	}
	for i := range padBuf {
		padBuf[i] = 0xFF
	}
	copy(padBuf[prefix:], p)
	_, err := rawBackend.WriteAt(padBuf[:total], pageStart)
	return len(p), err
}

var padBuf [512]byte

func (d flashBlockDev) Size() int64 {
	return flashStorageBytes
}

func (d flashBlockDev) WriteBlockSize() int64 {
	return writePage
}

func (d flashBlockDev) EraseBlockSize() int64 {
	return flashBlockSize
}

func (d flashBlockDev) EraseBlocks(start, length int64) error {
	if start < 0 || (start+length)*flashBlockSize > flashStorageBytes {
		return flashErrOutOfRange
	}
	return rawBackend.EraseBlocks(regionBase/flashBlockSize+start, length)
}