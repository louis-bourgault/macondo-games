//go:build wasm

package main

import (
	"time"

	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/game/menu"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/wasm"
)

func main() {
	//initiate the platform
	display := wasm.NewDisplay("canvas")
	input := wasm.NewInput()
	var currentGame game.Game
	currentGame = menu.New()

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	var lastTime = time.Now()

	for range ticker.C {
		now := time.Now()
		dt := now.Sub(lastTime)
		lastTime = now
		nextGame := currentGame.Update(dt.Seconds(), input)
		if nextGame == nil {
			currentGame = menu.New()
		} else {
			currentGame = nextGame
		}
		currentGame.Draw(display)
		input.Update()
	}
}
