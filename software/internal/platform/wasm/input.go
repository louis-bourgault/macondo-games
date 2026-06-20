package wasm

import (
	"syscall/js"

	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

type WasmInput struct {
	keysPressed      map[platform.Button]bool
	keysJustPressed  map[platform.Button]bool
	keysJustReleased map[platform.Button]bool
}

func NewInput() *WasmInput {
	input := &WasmInput{
		keysPressed:      make(map[platform.Button]bool),
		keysJustPressed:  make(map[platform.Button]bool),
		keysJustReleased: make(map[platform.Button]bool),
	}
	js.Global().Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyPressed := args[0].Get("key").String()
		var key platform.Button
		switch keyPressed {
		case "W", "w":
			key = platform.Up
		case "A", "a":
			key = platform.Left
		case "S", "s":
			key = platform.Down
		case "D", "d":
			key = platform.Right
		case "Enter":
			key = platform.Start
		case "Shift":
			key = platform.Exit
		case "j", "J":
			key = platform.A
		case "k", "K":
			key = platform.B
		}
		if !input.keysPressed[key] {
			input.keysJustPressed[key] = true
		}
		input.keysPressed[key] = true
		return nil
	}))
	js.Global().Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyPressed := args[0].Get("key").String()
		var key platform.Button
		switch keyPressed {
		case "W", "w":
			key = platform.Up
		case "A", "a":
			key = platform.Left
		case "S", "s":
			key = platform.Down
		case "D", "d":
			key = platform.Right
		case "Enter":
			key = platform.Start
		case "Shift":
			key = platform.Exit
		case "j", "J":
			key = platform.A
		case "k", "K":
			key = platform.B
		}
		if input.keysPressed[key] {
			input.keysJustReleased[key] = true
		}
		input.keysPressed[key] = false
		return nil
	}))
	return input
}

func (i *WasmInput) WasKeyJustPressed(key platform.Button) bool {
	return i.keysJustPressed[key]
}

func (i *WasmInput) WasKeyJustReleased(key platform.Button) bool {
	return i.keysJustReleased[key]
}

func (i *WasmInput) IsKeyPressed(key platform.Button) bool {
	return i.keysPressed[key]
}

func (i *WasmInput) Update() {
	// Clear just pressed/released states after each update
	for key := range i.keysJustPressed {
		delete(i.keysJustPressed, key)
	}
	for key := range i.keysJustReleased {
		delete(i.keysJustReleased, key)
	}
}
