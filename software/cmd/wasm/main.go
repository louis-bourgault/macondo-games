//go:build wasm

package main

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game/drawtest"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/wasm"
)

func main() {
	//initiate the platform
	display := wasm.NewDisplay("canvas")
	drawtest.New(display).Draw(display)
}
