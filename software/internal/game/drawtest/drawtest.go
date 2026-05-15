package drawtest

//an implementation of a simple game that draws a pixel.

import "github.com/louis-bourgault/macondo-games/software/internal/platform"

func New(s platform.Screen) *DrawTest {
	return &DrawTest{}
}

type DrawTest struct {
}

func (d *DrawTest) Update() {
	//no update logic for this test
}

func (d *DrawTest) Draw(screen platform.Screen) {
	screen.Pixel(10, 10, 0xF800) //draw a red pixel at (10, 10)
	screen.Present()             //present the screen
}
