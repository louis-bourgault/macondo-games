//go:build pico

package device

import (
	"machine"

	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

// Version one button wiring. The two small auxiliary buttons retain their
// existing software names for compatibility with the current board.
var hardwareButtonPins = [...]buttonPin{
	{platform.A, machine.Pin(4)},
	{platform.B, machine.Pin(5)},
	{platform.Start, machine.Pin(6)},
	{platform.Exit, machine.Pin(7)},
	{platform.Up, machine.Pin(9)},
	{platform.Left, machine.Pin(10)},
	{platform.Right, machine.Pin(11)},
	{platform.Down, machine.Pin(12)},
}
