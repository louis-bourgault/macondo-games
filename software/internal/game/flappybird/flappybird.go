package flappybird

//a template game to copy and build off.

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *FlappyBirdGame {
	return &FlappyBirdGame{}
}

type FlappyBirdGame struct {
}

func (d *FlappyBirdGame) Update(dt float64, input platform.InputSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	return d
}

func (d *FlappyBirdGame) Draw(screen platform.Screen) {
	screen.Fill(0x003f)
	helpers.DrawSprite(screen, BirdSprite, 10, 10)
	screen.Present() //present the screen
}
