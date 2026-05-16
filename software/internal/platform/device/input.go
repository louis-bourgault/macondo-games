package device

import "machine"

type HardwareInput struct {
	keysPressed      map[string]bool
	keysJustPressed  map[string]bool
	keysJustReleased map[string]bool
}

// gpois - look at the kicad schematic, change these if you're using a different schematic
const (
	PIN_A      = machine.PIN(4)
	PIN_B      = machine.PIN(5)
	PIN_SELECT = machine.PIN(6)
	PIN_START  = machine.PIN(7)
	DPAD_UP    = machine.PIN(9)
	DPAD_LEFT  = machine.PIN(10)
	DPAD_RIGHT = machine.PIN(11)
	DPAD_DOWN  = machine.PIN(12)
)

func NewHardwareInput() *HardwareInput {
	//since the other side of the pin is gnd, we need internal pullups.
	PIN_A.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	PIN_B.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	PIN_SELECT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	PIN_START.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_UP.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_LEFT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_RIGHT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	DPAD_DOWN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	return &HardwareInput{
		keysPressed:      make(map[string]bool),
		keysJustPressed:  make(map[string]bool),
		keysJustReleased: make(map[string]bool),
	}
}

func (i *HardwareInput) WasKeyJustPressed(key string) bool {
	//pretty basic function tbh
	return i.keysJustPressed[key]
}

func (i *HardwareInput) WasKeyJustReleased(key string) bool {
	return i.keysJustReleased[key]
}

func (i *HardwareInput) IsKeyPressed(key string) bool {
	return i.keysPressed[key]
}

func (i *HardwareInput) Update() {
	i.updateKey("A", PIN_A)
	i.updateKey("B", PIN_B)
	i.updateKey("SELECT", PIN_SELECT)
	i.updateKey("START", PIN_START)
	i.updateKey("UP", DPAD_UP)
	i.updateKey("LEFT", DPAD_LEFT)
	i.updateKey("RIGHT", DPAD_RIGHT)
	i.updateKey("DOWN", DPAD_DOWN)

	//clear jp/jr states after each update
	for key := range i.keysJustPressed {
		delete(i.keysJustPressed, key)
	}
	for key := range i.keysJustReleased {
		delete(i.keysJustReleased, key)
	}
}

func (i *HardwareInput) updateKey(key string, pin machine.Pin) {
	pressed := !pin.Get() //active low, bc of pullups
	if pressed && !i.keysPressed[key] {
		i.keysJustPressed[key] = true
	} else if !pressed && i.keysPressed[key] {
		i.keysJustReleased[key] = true
	}
	i.keysPressed[key] = pressed
}
