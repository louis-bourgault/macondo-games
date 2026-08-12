// bridge.go exposes the Ferretboard engine to MicroPython.
//
// Each //export func is a C-callable symbol that the host.c registers as a
// `ferret` module function. They act on the singleton `display` (see main.go)
// and the shared helpers package. This file is the ONLY place student-facing
// SDK calls are mapped to Go; keep it thin.
//
// Argument convention:
//   - ints/uint16 marshaled as plain C.int / C.uint16_t (no struct).
//   - strings marshaled as *C.char; convert with C.GoString.

package main

/*
#include <stdint.h>
*/
import "C"

import (
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
)

//export ferret_fill
func ferret_fill(color C.uint16_t) {
	display.Fill(uint16(color))
}

//export ferret_fill_rect
func ferret_fill_rect(x, y, w, h C.int, color C.uint16_t) {
	display.FillRect(int(x), int(y), int(w), int(h), uint16(color))
}

//export ferret_pixel
func ferret_pixel(x, y C.int, color C.uint16_t) {
	display.Pixel(int(x), int(y), uint16(color))
}

//export ferret_present
func ferret_present() {
	display.Present()
}

//export ferret_draw_text
func ferret_draw_text(text *C.char, x, y C.int, color C.uint16_t) {
	helpers.DrawText(display, int(x), int(y), C.GoString(text), uint16(color))
}

//export ferret_random_int
func ferret_random_int(min, max C.int) C.int {
	return C.int(helpers.RandomInt(int(min), int(max)))
}
