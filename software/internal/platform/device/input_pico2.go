//go:build pico2

package device

import (
	"machine"

	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

// Version two button wiring from hardware-v2. V2 replaces Start/Exit with a
// single Menu button and adds X/Y alongside A/B.
var hardwareButtonPins = [...]buttonPin{
	{platform.A, machine.Pin(4)},
	{platform.B, machine.Pin(5)},
	{platform.Menu, machine.Pin(6)},
	{platform.Up, machine.Pin(9)},
	{platform.Left, machine.Pin(10)},
	{platform.Right, machine.Pin(11)},
	{platform.Down, machine.Pin(12)},
	{platform.X, machine.Pin(13)},
	{platform.Y, machine.Pin(14)},
}
