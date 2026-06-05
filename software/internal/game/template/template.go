package template

//a template game to copy and build off.

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *TemplateGame {
	return &TemplateGame{}
}

type TemplateGame struct {
}

func (d *TemplateGame) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	return d
}

func (d *TemplateGame) Draw(screen platform.Screen) {
	screen.Fill(0x003f)
	screen.Pixel(120, 120, 0xF800) //draw a red pixel at (120,120)
	screen.Present()               //present the screen
}
