//go:build tinygo

package main

/*
#include <stdint.h>
*/
import "C"
import (
	"machine"
	"runtime"
	"unsafe"
)

// CDC ring buffer: Go's USB read pump fills it; MP drains it via ferret_cdc_read.
//
// Sized for the web editor's largest paste block (a base64 image chunk). The
// pump must never overwrite unread bytes: on the RP2040's cooperative
// scheduler the pump can run far ahead of the REPL while MP is busy echoing a
// paste, and a too-small buffer silently dropped the wrapper line
// (import ferret + ferret.write_image(...)), leaving only raw base64 to be
// executed. Keep the buffer big enough for one full paste and backpressure
// (yield, don't overwrite) when it fills.
var cdcBuf [4096]byte
var cdcHead, cdcTail int // head = write, tail = read

//export ferret_cdc_read
func ferret_cdc_read() C.int {
	for cdcHead == cdcTail {
		drainCDC()
		runtime.Gosched()
	}
	b := cdcBuf[cdcTail]
	cdcTail = (cdcTail + 1) % len(cdcBuf)
	// Drain on the way out too: MP can chew through the ring faster than the
	// pump goroutine gets scheduled, and a full TinyUSB RX ring (512 B) drops
	// bytes, not backpressure.
	drainCDC()
	return C.int(b)
}

//export ferret_cdc_write
func ferret_cdc_write(s *C.char, n C.int) {
	if n <= 0 {
		return
	}
	// TinyGo's USBCDC.Write silently drops bytes until the host has opened the
	// port and asserted DTR, so the boot banner + ">>> " prompt would vanish if
	// nobody is attached yet. Wait for the terminal first: the user sees the
	// real MicroPython banner the instant they connect instead of a blank
	// window.
	for !machine.Serial.DTR() {
		runtime.Gosched()
	}
	// Treat s as a byte slice of length n and write to USB CDC.
	buf := (*[1 << 20]byte)(unsafe.Pointer(s))[:n:n]
	machine.Serial.Write(buf)
}

// pumpCDC reads from USB CDC into the ring buffer. Runs as a goroutine so the
// REPL loop in host.c can poll ferret_cdc_read without blocking on USB I/O.
func drainCDC() {
	for machine.Serial.Buffered() > 0 {
		next := (cdcHead + 1) % len(cdcBuf)
		if next == cdcTail {
			// Ring is full; the REPL hasn't drained it yet. Stop so we don't
			// overwrite unread bytes (which silently corrupts large pastes).
			break
		}
		if b, err := machine.Serial.ReadByte(); err == nil {
			cdcBuf[cdcHead] = b
			cdcHead = next
		}
	}
}

func pumpCDC() {
	for {
		drainCDC()
		// Must yield every iteration: the RP2040 scheduler is cooperative, so a
		// busy loop without a scheduling point starves the REPL goroutine in
		// ferret_cdc_read and the device never responds.
		runtime.Gosched()
	}
}
