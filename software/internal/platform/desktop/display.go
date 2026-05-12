//go:build !device

package desktop

import (
	"image"
	"image/color"
	"image/draw"
)

type Display struct {
	buf *image.RGBA
}

func NewDisplay(width, height int) *Display {
	return &Display{
		buf: image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

func (d *Display) FillScreen(c color.Color) {
	draw.Draw(d.buf, d.buf.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
}

func (d *Display) DrawPixel(x, y int, c color.Color) {
	d.buf.Set(x, y, c)
}

func (d *Display) DrawRect(x, y, w, h int, c color.Color, fill bool) {
	if fill {
		draw.Draw(d.buf, image.Rect(x, y, x+w, y+h), &image.Uniform{c}, image.Point{}, draw.Src)
	} else {
		for i := x; i < x+w; i++ {
			d.buf.Set(i, y, c)
			d.buf.Set(i, y+h-1, c)
		}
		for j := y; j < y+h; j++ {
			d.buf.Set(x, j, c)
			d.buf.Set(x+w-1, j, c)
		}
	}
}

func (d *Display) Present() error {
	//integrate to draw it to a window on the desktop
	return nil
}
func (d *Display) Buffer() *image.RGBA {
	return d.buf
}
