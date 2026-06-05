package wasm

import "syscall/js"

func Log(s string) {
	js.Global().Get("console").Call("log", s)
}

type WasmLogSystem struct {
}

func NewLogSystem() *WasmLogSystem {
	return &WasmLogSystem{}
}

func (l *WasmLogSystem) Log(s string) {
	Log(s)
}
