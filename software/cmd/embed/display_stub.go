//go:build !tinygo

// Host reference stub for the display singleton. Used when building the embed
// target on a PC (no SPI hardware) so the reference links and runs. On the
// device, display_device.go (tagged `tinygo`) provides the real ST7789 display.

//This is primarily useful for headless testing, which is neccesary because
//a lot of this code was written by agents without hardware access.


package main

import "github.com/louis-bourgault/macondo-games/software/internal/platform"

// stubDisplay is a no-op implementation of platform.Screen for the host build.
type stubDisplay struct{}

func (s *stubDisplay) Fill(c uint16)                     {}
func (s *stubDisplay) FillRect(x, y, w, h int, c uint16) {}
func (s *stubDisplay) Pixel(x, y int, c uint16)          {}
func (s *stubDisplay) Present() error                    { return nil }

// newDisplay returns the host stub.
func newDisplay() platform.Screen {
	return &stubDisplay{}
}
