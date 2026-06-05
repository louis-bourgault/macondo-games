package main

import (
	"time"

	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/game/menu"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/device"
)

func main() {
	//initiate the platform
	//perhaps we should seed the randomness here
	display := device.NewDisplay()
	input := device.NewHardwareInput()
	var currentGame game.Game
	currentGame = menu.New()
	log := device.NewLogSystem()
	currentGame = menu.New()

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	var lastTime = time.Now()

	for range ticker.C {
		now := time.Now()
		dt := now.Sub(lastTime)
		lastTime = now
		nextGame := currentGame.Update(dt.Seconds(), input, log)
		if nextGame == nil {
			currentGame = menu.New()
		} else {
			currentGame = nextGame
		}
		currentGame.Draw(display)
		input.Update()
	}
}
