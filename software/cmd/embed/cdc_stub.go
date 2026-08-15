//go:build !tinygo

package main

/*
#include <stdint.h>
*/
import "C"

import (
	"os"
	"unsafe"
)

// Host-reference CDC hooks. On a PC there is no USB serial, so the stub
// bridges to the real terminal: reads come from stdin and writes go to stdout.
// This makes `just host` an interactive MP REPL with no hardware, and lets the
// MP side (mp_hal_stdout_tx_strn, run_repl echo) behave as it would on device.

//export ferret_cdc_read
func ferret_cdc_read() C.int {
	var b [1]byte
	n, _ := os.Stdin.Read(b[:])
	if n == 0 {
		return -1 // empty / EOF
	}
	return C.int(b[0])
}

//export ferret_cdc_write
func ferret_cdc_write(s *C.char, n C.int) {
	if n <= 0 {
		return
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(s)), int(n))
	_, _ = os.Stdout.Write(buf)
}

//export ferret_cdc_poll_interrupt
func ferret_cdc_poll_interrupt(interruptChar C.int) C.int {
	// The host stub blocks directly in ferret_cdc_read, so there is no
	// non-blocking input source to poll while bytecode is executing.
	return 0
}
