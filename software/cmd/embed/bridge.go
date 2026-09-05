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
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
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

//export ferret_input_update
func ferret_input_update() {
	input.Update()
}

// buttonFromString maps the MicroPython SDK key names onto the platform.Button
// enum, matching the standalone SDK's string API ("A", "B", "UP", "DOWN",
// "LEFT", "RIGHT", plus the board-specific auxiliary keys). Unknown names
// read as platform.None, i.e. always unpressed.
func buttonFromString(name string) platform.Button {
	switch name {
	case "UP":
		return platform.Up
	case "DOWN":
		return platform.Down
	case "LEFT":
		return platform.Left
	case "RIGHT":
		return platform.Right
	case "A":
		return platform.A
	case "B":
		return platform.B
	case "START":
		return platform.Start
	case "EXIT":
		return platform.Exit
	case "X":
		return platform.X
	case "Y":
		return platform.Y
	case "MENU":
		return platform.Menu
	default:
		return platform.None
	}
}

func boolToCInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

//export ferret_input_is_pressed
func ferret_input_is_pressed(key *C.char) C.int {
	return boolToCInt(input.IsKeyPressed(buttonFromString(C.GoString(key))))
}

//export ferret_input_was_just_pressed
func ferret_input_was_just_pressed(key *C.char) C.int {
	return boolToCInt(input.WasKeyJustPressed(buttonFromString(C.GoString(key))))
}

//export ferret_input_was_just_released
func ferret_input_was_just_released(key *C.char) C.int {
	return boolToCInt(input.WasKeyJustReleased(buttonFromString(C.GoString(key))))
}

//export ferret_rgb_to_565
func ferret_rgb_to_565(r, g, b C.int) C.uint16_t {
	return C.uint16_t(helpers.RGBto565(int(r), int(g), int(b)))
}

//export ferret_measure_text
func ferret_measure_text(text *C.char, w, h *C.int) {
	wv, hv := helpers.MeasureText(C.GoString(text))
	*w = C.int(wv)
	*h = C.int(hv)
}
