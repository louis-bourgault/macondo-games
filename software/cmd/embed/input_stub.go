//go:build !tinygo

// Host reference stub for the input singleton. Used when building the embed
// target on a PC (no button hardware) so the reference links and runs. All
// keys read as released. On the device, input_device.go (tagged `tinygo`)
// provides the real button reading.

package main

import "github.com/louis-bourgault/macondo-games/software/internal/platform"

// stubInput is a no-op implementation of platform.InputSystem for the host build.
type stubInput struct{}

func (s *stubInput) WasKeyJustPressed(key platform.Button) bool  { return false }
func (s *stubInput) WasKeyJustReleased(key platform.Button) bool { return false }
func (s *stubInput) IsKeyPressed(key platform.Button) bool      { return false }
func (s *stubInput) Update()                                    {}

// newInput returns the host stub.
func newInput() platform.InputSystem {
	return &stubInput{}
}