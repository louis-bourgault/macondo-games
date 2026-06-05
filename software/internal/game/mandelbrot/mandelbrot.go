package mandelbrot

//a template game to copy and build off.

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *Mandelbrot {
	return &Mandelbrot{
		Maxiter: 1,
		draw:    true,
		ItrMode: true,
		zoom:    0.015625,
		offset:  coord{x: 120, y: 120},
		IterMax: 50,
	}
}

type coord struct {
	x, y int
}

type Mandelbrot struct {
	Maxiter int
	draw    bool
	ItrMode bool
	zoom    float32
	IterMax int
	offset  coord
}

func (d *Mandelbrot) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	if input.WasKeyJustPressed(platform.Start) {
		d.ItrMode = !d.ItrMode

	}
	d.draw = false
	if d.ItrMode {
		if input.WasKeyJustPressed(platform.B) && d.Maxiter < d.IterMax {
			d.Maxiter += 1
			d.draw = true
		}
		if input.WasKeyJustPressed(platform.A) && d.Maxiter > 0 {
			d.Maxiter -= 1
			d.draw = true
		}
	} else {
		if input.WasKeyJustPressed(platform.A) {
			// Zoom in - keep center point at same complex value
			centerX := float32(120-d.offset.x) * d.zoom
			centerY := float32(120-d.offset.y) * d.zoom
			d.zoom = d.zoom / 1.5
			d.offset.x = 120 - int(centerX/d.zoom)
			d.offset.y = 120 - int(centerY/d.zoom)
			d.draw = true
		}
		if input.WasKeyJustPressed(platform.B) {
			// Zoom out - keep center point at same complex value
			centerX := float32(120-d.offset.x) * d.zoom
			centerY := float32(120-d.offset.y) * d.zoom
			d.zoom = d.zoom * 1.5
			d.offset.x = 120 - int(centerX/d.zoom)
			d.offset.y = 120 - int(centerY/d.zoom)
			d.draw = true
		}

		if input.IsKeyPressed(platform.Up) {
			d.offset.y += 1
			d.draw = true
		}
		if input.IsKeyPressed(platform.Down) {
			d.offset.y -= 1
			d.draw = true
		}
		if input.IsKeyPressed(platform.Left) {
			d.offset.x += 1
			d.draw = true
		}
		if input.IsKeyPressed(platform.Right) {
			d.offset.x -= 1
			d.draw = true
		}
	}

	return d
}

func (d *Mandelbrot) Draw(screen platform.Screen) {
	if d.draw == false {
		return
	}
	for x := 0; x < 240; x++ {
		for y := 0; y < 240; y++ {
			c := complex(float32(x-d.offset.x)*d.zoom, float32(y-d.offset.y)*d.zoom)
			n := mandelIter(0, c, d.Maxiter)
			col := helpers.RGBto565(int(n*5), int(n*2), int(n*10))
			screen.Pixel(x, y, col)
		}
	}
	screen.Present()

}
