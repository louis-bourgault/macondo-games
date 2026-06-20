package drawtest

//an implementation of a simple game that draws a pixel.
//i know, riveting

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *DrawTest {
	return &DrawTest{
		timeSince: 0,
		xp:        120,
		yp:        120,
	}
}

type DrawTest struct {
	timeSince int
	xp        int
	yp        int
}

func (d *DrawTest) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	//d.timeSince++
	if input.WasKeyJustPressed(platform.Right) {
		d.xp += 10
	}
	if input.WasKeyJustPressed(platform.Left) {
		d.xp -= 10
	}
	if input.WasKeyJustPressed(platform.Up) {
		d.yp -= 10
	}
	if input.WasKeyJustPressed(platform.Down) {
		d.yp += 10
	}

	if input.WasKeyJustPressed(platform.Exit) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	return d
}

func (d *DrawTest) Draw(screen platform.Screen) {
	screen.Fill(0x003f)
	screen.Pixel(d.xp, d.yp, 0xF800) //draw a red pixel at (d.xp, d.yp)
	helpers.DrawText(screen, 5, 5, "drawtest", 0xF800)
	screen.Present() //present the screen
}
