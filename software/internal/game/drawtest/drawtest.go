package drawtest

//an implementation of a simple game that draws a pixel.

import (
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New(s platform.Screen) *DrawTest {
	return &DrawTest{
		timeSince: 0,
		xp:        0,
	}
}

type DrawTest struct {
	timeSince int
	xp        int
}

func (d *DrawTest) Update(input platform.InputSystem) {
	//d.timeSince++
	if input.WasKeyJustPressed("RIGHT") {
		d.xp += 10
	}
}

func (d *DrawTest) Draw(screen platform.Screen) {
	screen.Pixel(d.xp, 10, 0xF800) //draw a red pixel at (d.xp, 10)
	screen.Present()               //present the screen
}
