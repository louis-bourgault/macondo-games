//go:build !tinygo

package main

import (
	"errors"
	"os"
	"path/filepath"
)

// flashErrOutOfRange mirrors the tinygo flash driver's range error.
var flashErrOutOfRange = errors.New("flash: read or write beyond filesystem region")

// stubbedFlash models the 16 MiB flash chip as a plain file, per the same
// geometry as the device: the filesystem region lives at chip offset 8 MiB.
type stubbedFlash struct {
	f *os.File
}

const hostFlashSize = 16 * 1024 * 1024

func setupFlashRegion() {
	regionBase = hostFlashSize - flashStorageBytes
	rawTotalSize = hostFlashSize

	path := os.Getenv("FERRET_FS_IMAGE")
	if path == "" {
		path = filepath.Join("build", "ferret.img")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	st, err := f.Stat()
	if err != nil {
		panic(err)
	}
	if st.Size() < hostFlashSize {
		// expand the image; new bytes read as 0x00 on a file, but flash
		// reads erased sectors as 0xFF, so initialise explicitly.
		if err := f.Truncate(hostFlashSize); err != nil {
			panic(err)
		}
		fill := make([]byte, 64*1024)
		for i := range fill {
			fill[i] = 0xFF
		}
		for off := int64(0); off < hostFlashSize; off += int64(len(fill)) {
			if _, err := f.WriteAt(fill, off); err != nil {
				panic(err)
			}
		}
	}
	rawBackend = stubbedFlash{f: f}
}

func (s stubbedFlash) ReadAt(p []byte, off int64) (int, error) {
	return s.f.ReadAt(p, off)
}

// WriteAt emulates NOR-flash page programming: programming only ever clears
// bits, so the resulting bytes are the bitwise AND of the existing content
// and the written data. A plain file overwrite would clobber bytes that were
// programmed into the same page by an earlier prog call.
func (s stubbedFlash) WriteAt(p []byte, off int64) (int, error) {
	old := make([]byte, len(p))
	if _, err := s.f.ReadAt(old, off); err != nil {
		return 0, err
	}
	for i := range p {
		p[i] &= old[i]
	}
	return s.f.WriteAt(p, off)
}

func (s stubbedFlash) EraseBlocks(start, length int64) error {
	off := start * flashBlockSize
	fill := make([]byte, flashBlockSize)
	for i := range fill {
		fill[i] = 0xFF
	}
	for b := int64(0); b < length; b++ {
		if _, err := s.f.WriteAt(fill, off+b*flashBlockSize); err != nil {
			return err
		}
	}
	return nil
}

func (s stubbedFlash) Size() int64 { return hostFlashSize }