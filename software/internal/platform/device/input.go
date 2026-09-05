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

type buttonPin struct {
	button platform.Button
	pin    machine.Pin
}

func NewHardwareInput() *HardwareInput {
	for _, key := range hardwareButtonPins {
		key.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

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
	//clear before updating, not after.
	for key := range i.keysJustPressed {
		delete(i.keysJustPressed, key)
	}
	for key := range i.keysJustReleased {
		delete(i.keysJustReleased, key)
	}
	for _, key := range hardwareButtonPins {
		i.updateKey(key.button, key.pin)
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
