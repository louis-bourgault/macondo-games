package main

//simple test.
//ai disclosure: fully made by deepseek v4 flash with opencode harness, since its just a test not the actual code.

import (
	"time"

	"github.com/louis-bourgault/macondo-games/software/internal/platform/device"
)

func main() {
	time.Sleep(2 * time.Second)
	println("Starting display test...")

	display := device.NewDisplay()
	println("Display initialized")

	// Main loop cycling through different colors to verify display rendering
	colors := []uint16{
		0xF800, // Red
		0x07E0, // Green
		0x001F, // Blue
		0xFFFF, // White
		0x0000, // Black
	}

	colorIdx := 0
	for {
		color := colors[colorIdx]
		println("Filling screen with color:", color)

		// Fill the entire buffer
		display.Fill(color)

		// Draw some contrasting rectangles to test coordinates and boundary limits
		display.FillRect(10, 10, 50, 50, 0xFFFF-color)
		display.FillRect(180, 180, 50, 50, 0xFFFF-color)

		// Present to display
		err := display.Present()
		if err != nil {
			println("Present error:", err.Error())
		}

		colorIdx = (colorIdx + 1) % len(colors)
		time.Sleep(1 * time.Second)
	}
}
