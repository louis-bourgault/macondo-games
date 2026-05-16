package wasm

import "syscall/js"

type WasmInput struct {
	keysPressed      map[string]bool
	keysJustPressed  map[string]bool
	keysJustReleased map[string]bool
}

func NewInput() *WasmInput {
	input := &WasmInput{
		keysPressed:      make(map[string]bool),
		keysJustPressed:  make(map[string]bool),
		keysJustReleased: make(map[string]bool),
	}
	js.Global().Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		key := args[0].Get("key").String()
		switch key {
		case "W", "w":
			key = "UP"
		case "A", "a":
			key = "LEFT"
		case "S", "s":
			key = "DOWN"
		case "D", "d":
			key = "RIGHT"
		case "Enter":
			key = "START"
		case "Shift":
			key = "SELECT"
		case "j", "J":
			key = "A"
		case "k", "K":
			key = "B"
		}
		if !input.keysPressed[key] {
			input.keysJustPressed[key] = true
		}
		input.keysPressed[key] = true
		return nil
	}))
	js.Global().Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		key := args[0].Get("key").String()
		switch key {
		case "W", "w":
			key = "UP"
		case "A", "a":
			key = "LEFT"
		case "S", "s":
			key = "DOWN"
		case "D", "d":
			key = "RIGHT"
		case "Enter":
			key = "START"
		case "Shift":
			key = "SELECT"
		case "j", "J":
			key = "A"
		case "k", "K":
			key = "B"
		}
		if input.keysPressed[key] {
			input.keysJustReleased[key] = true
		}
		input.keysPressed[key] = false
		return nil
	}))
	return input
}

func (i *WasmInput) WasKeyJustPressed(key string) bool {
	return i.keysJustPressed[key]
}

func (i *WasmInput) WasKeyJustReleased(key string) bool {
	return i.keysJustReleased[key]
}

func (i *WasmInput) IsKeyPressed(key string) bool {
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
