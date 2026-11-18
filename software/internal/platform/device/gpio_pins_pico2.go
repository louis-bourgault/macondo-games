//go:build pico2

package device

import (
	"machine"
)

var gpioAvailable = true

var gpioPins = [...]buttonPin{
	{1, machine.Pin(29)},
	{2, machine.Pin(28)},
	{3, machine.Pin(27)},
	{4, machine.Pin(26)},
}

// this pin can tell if 5v is connected. its not reccomended to run gpio off battery
var usbDetectPin = machine.Pin(8)
