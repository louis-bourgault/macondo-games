//go:build wasm

package main

import (
	"time"

	"github.com/louis-bourgault/macondo-games/software/internal/game/drawtest"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/wasm"
)

func main() {
	//initiate the platform
	display := wasm.NewDisplay("canvas")
	input := wasm.NewInput()
	game := drawtest.New(display)

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for range ticker.C {
		game.Update(input)
		game.Draw(display)
	}
}
