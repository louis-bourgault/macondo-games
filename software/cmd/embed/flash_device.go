//go:build tinygo

package main

import (
	"machine"
)

// flashErrOutOfRange mirrors the tinygo flash driver's range error.
var flashErrOutOfRange = errFlashRange

// setupFlashRegion is called from main() (not init, so TinyGo's compile-time
// interp never tries to fold the machine.FlashDataStart/End addresses; the
// interp panics on the pointer-to-int conversion of the __flash_data_end
// symbol).
func setupFlashRegion() {
	// The FS region is the last 8 MiB of the flash chip. TinyGo's
	// machine.Flash exposes the whole chip (__flash_size=16M in
	// ferretboard.json) as offset 0 = FlashDataStart, so the region base is
	// measured from there.
	start := int64(machine.FlashDataStart())
	end := int64(machine.FlashDataEnd())
	rawTotalSize = end - start
	regionBase = rawTotalSize - flashStorageBytes
	if regionBase < 0 {
		// Paranoia: if the firmware ever outgrows the first 8 MiB this
		// shifts the region down instead of clamping to nothing.
		panic("firmware larger than first 8 MiB of flash")
	}
	rawBackend = machine.Flash
}

var errFlashRange = flashRangeError{}

type flashRangeError struct{}

func (flashRangeError) Error() string { return "flash: read or write beyond filesystem region" }