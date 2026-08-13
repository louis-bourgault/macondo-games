// flash_common.go — the Ferretboard flash block device.
//
// layout: the firmware programs live in the first part of the 16 MiB W25Q128
// QSPI flash; the last 8 MiB are reserved for the user filesystem ("assets and
// user code"). That region is presented to both runtimes as a block device
// with blocks of 4096 bytes (the flash erase-sector size):
//
//   - MicroPython (port/flash_storage.c) talks to it through the //export
//     funcs below, wrapped up as an MP "Flash" object with readblocks/
//     writeblocks/ioctl; host.c mounts it as "/" at boot.
//   - Go code (flash_fs.go) talks to it through the same flashBlockDev,
//     mounted with tinygo.org/x/tinyfs, so both runtimes read and write the
//     exact same LittleFS2 volume.
//
// The raw I/O is per-build: on the device it is machine.Flash (TinyGo's
// driver for the RP2040/RP2350 bootrom flash routines); on the host reference
// it is a plain-file image (flash_stub.go).

package main

/*
#include <stdint.h>
*/
import "C"

import (
	"unsafe"
)

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
// region's offset from its consumers (littlefs2 in C on the MP side via the
// exports, and tinyfs on the Go side).
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

// --- the //export funcs MicroPython's Flash object calls --------------------

//export ferret_flash_read
func ferret_flash_read(block C.uint32_t, off C.uint32_t, buf *C.uint8_t, n C.uint32_t) C.int {
	p := unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(n))
	if _, err := (flashBlockDev{}).ReadAt(p, int64(block)*flashBlockSize+int64(off)); err != nil {
		return -1
	}
	return 0
}

//export ferret_flash_write
func ferret_flash_write(block C.uint32_t, off C.uint32_t, buf *C.uint8_t, n C.uint32_t) C.int {
	p := unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(n))
	if _, err := (flashBlockDev{}).WriteAt(p, int64(block)*flashBlockSize+int64(off)); err != nil {
		return -1
	}
	return 0
}

//export ferret_flash_erase
func ferret_flash_erase(block C.uint32_t) C.int {
	if err := (flashBlockDev{}).EraseBlocks(int64(block), 1); err != nil {
		return -1
	}
	return 0
}

//export ferret_flash_block_count
func ferret_flash_block_count() C.int {
	return C.int(flashRegionBlocks)
}

//export ferret_flash_block_size
func ferret_flash_block_size() C.int {
	return C.int(flashBlockSize)
}