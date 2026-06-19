package device

import (
	"machine"

	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

type HardwareInput struct {
	KeysPressed      map[platform.Button]bool
	keysJustPressed  map[platform.Button]bool
	keysJustReleased map[platform.Button]bool
}

// gpois - look at the kicad schematic, change these if you're using a different schematic
const (
	PIN_A      = machine.Pin(4)
	PIN_B      = machine.Pin(5)
	PIN_SELECT = machine.Pin(6)
	PIN_START  = machine.Pin(7)
	DPAD_UP    = machine.Pin(9)
	DPAD_LEFT  = machine.Pin(10)
	DPAD_RIGHT = machine.Pin(11)
	DPAD_DOWN  = machine.Pin(12)
)

func NewHardwareInput() *HardwareInput {
	PIN_A.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	PIN_B.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	PIN_SELECT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	PIN_START.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_UP.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_LEFT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_RIGHT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_DOWN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	return &HardwareInput{
		KeysPressed:      make(map[platform.Button]bool),
		keysJustPressed:  make(map[platform.Button]bool),
		keysJustReleased: make(map[platform.Button]bool),
	}
}

func (i *HardwareInput) WasKeyJustPressed(key platform.Button) bool {
	//pretty basic function tbh
	return i.keysJustPressed[key]
}

func (i *HardwareInput) WasKeyJustReleased(key platform.Button) bool {
	return i.keysJustReleased[key]
}

func (i *HardwareInput) IsKeyPressed(key platform.Button) bool {
	return i.KeysPressed[key]
}

func (i *HardwareInput) Update() {
	i.updateKey(platform.A, PIN_A)
	i.updateKey(platform.B, PIN_B)
	i.updateKey(platform.Select, PIN_SELECT)
	i.updateKey(platform.Start, PIN_START)
	i.updateKey(platform.Up, DPAD_UP)
	i.updateKey(platform.Left, DPAD_LEFT)
	i.updateKey(platform.Right, DPAD_RIGHT)
	i.updateKey(platform.Down, DPAD_DOWN)

	//clear jp/jr states after each update
	for key := range i.keysJustPressed {
		delete(i.keysJustPressed, key)
	}
	for key := range i.keysJustReleased {
		delete(i.keysJustReleased, key)
	}
}

func (i *HardwareInput) updateKey(key platform.Button, pin machine.Pin) {
	pressed := !pin.Get() //active low, bc of pullups
	if pressed && !i.KeysPressed[key] {
		i.keysJustPressed[key] = true
	} else if !pressed && i.KeysPressed[key] {
		i.keysJustReleased[key] = true
	}
	i.KeysPressed[key] = pressed
}
