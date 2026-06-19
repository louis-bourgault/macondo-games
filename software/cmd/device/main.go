package main

import (
	"math/rand"
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
	randSeed := false

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	var lastTime = time.Now()
	var randTick int64
	randTick = 0

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
		if !randSeed {
			randTick++
			if len(input.KeysPressed) > 0 {
				//tinygo needs randomness to be seeded. we don't ahve a radnom source, so we use the n of ticks till the user does something.
				rand.Seed(int64(randTick))
				randSeed = true
			}
		}
	}
}
