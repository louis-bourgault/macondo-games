//go:build tinygo

// Device display wiring for the embed target. Uses the real Ferretboard
// hardware display (ST7789 over SPI) via the existing platform/device package.
// Compiled only for the `tinygo` build (Pico/RP2350).

package main

import (
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/device"
)

// newDisplay returns the real hardware display for the device build.
func newDisplay() platform.Screen {
	return device.NewDisplay()
}

// init starts the USB CDC read pump (TinyGo only) so the REPL loop in host.c
// can drain buffered bytes via ferret_cdc_read.
func init() {
	go pumpCDC()
}
