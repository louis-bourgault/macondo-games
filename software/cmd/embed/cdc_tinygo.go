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
var cdcBuf [512]byte
var cdcHead, cdcTail int // head = write, tail = read

//export ferret_cdc_read
func ferret_cdc_read() C.int {
	for cdcHead == cdcTail {
		runtime.Gosched()
	}
	b := cdcBuf[cdcTail]
	cdcTail = (cdcTail + 1) % len(cdcBuf)
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
func pumpCDC() {
	for {
		for machine.Serial.Buffered() > 0 {
			if b, err := machine.Serial.ReadByte(); err == nil {
				cdcBuf[cdcHead] = b
				cdcHead = (cdcHead + 1) % len(cdcBuf)
			}
		}
		// Must yield every iteration: the RP2040 scheduler is cooperative, so a
		// busy loop without a scheduling point starves the REPL goroutine in
		// ferret_cdc_read and the device never responds.
		runtime.Gosched()
	}
}
