package julia

//a template game to copy and build off.

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *Julia {
	return &Julia{
		Maxiter: 1,
		draw:    true,
		c:       coord{x: 0, y: 0},
		IterMax: 200,
	}
}

type coord struct {
	x, y int
}

type Julia struct {
	Maxiter int
	draw    bool
	c       coord
	IterMax int
}

func (d *Julia) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Exit) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}

	d.draw = false

	if input.WasKeyJustPressed(platform.B) && d.Maxiter < d.IterMax {
		d.Maxiter += 10
		d.draw = true
	}
	if input.WasKeyJustPressed(platform.A) && d.Maxiter > 0 {
		d.Maxiter -= 10
		d.draw = true
	}

	if input.IsKeyPressed(platform.Up) {
		d.c.y -= 1
		d.draw = true
	}
	if input.IsKeyPressed(platform.Down) {
		d.c.y += 1
		d.draw = true
	}
	if input.IsKeyPressed(platform.Left) {
		d.c.x -= 1
		d.draw = true
	}
	if input.IsKeyPressed(platform.Right) {
		d.c.x += 1
		d.draw = true
	}

	return d
}

func (d *Julia) Draw(screen platform.Screen) {
	if d.draw == false {
		return
	}
	for x := 0; x < 240; x++ {
		for y := 0; y < 240; y++ {
			n := juliaIter(complex(float32(x-120)/100, float32(y-120)/100), complex(float32(d.c.x-120)/100, float32(d.c.y-120)/100), d.Maxiter)
			col := helpers.RGBto565(int(n*5), int(n*2), int(n*10))
			screen.Pixel(x, y, col)
		}
	}
	screen.Pixel(d.c.x, d.c.y, helpers.RGBto565(255, 255, 255))
	screen.Present()

}
