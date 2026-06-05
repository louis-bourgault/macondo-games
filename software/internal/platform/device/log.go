package device

func Log(s string) {
	//do nothing
}

type WasmLogSystem struct {
}

func NewLogSystem() *WasmLogSystem {
	return &WasmLogSystem{}
}

func (l *WasmLogSystem) Log(s string) {
	Log(s)
}
