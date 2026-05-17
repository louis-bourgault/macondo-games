package drawtest

//an implementation of a simple game that draws a pixel.
//i know, riveting

import (
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New(s platform.Screen) *DrawTest {
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

func (d *DrawTest) Update(input platform.InputSystem) {
	//d.timeSince++
	if input.WasKeyJustPressed("RIGHT") {
		d.xp += 10
	}
	if input.WasKeyJustPressed("LEFT") {
		d.xp -= 10
	}
	if input.WasKeyJustPressed("UP") {
		d.yp -= 10
	}
	if input.WasKeyJustPressed("DOWN") {
		d.yp += 10
	}
}

func (d *DrawTest) Draw(screen platform.Screen) {
	screen.Fill(0x003f)
	screen.Pixel(d.xp, d.yp, 0xF800) //draw a red pixel at (d.xp, d.yp)
	screen.Present()                 //present the screen
}
