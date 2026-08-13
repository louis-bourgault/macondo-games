//go:build tinygo

// Device input wiring for the embed target. Uses the real Ferretboard buttons
// (active-low with internal pull-ups) via the existing platform/device package.
// Compiled only for the `tinygo` build (Pico/RP2350).

package main

import (
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/device"
)

// newInput returns the real hardware input for the device build.
func newInput() platform.InputSystem {
	return device.NewHardwareInput()
}