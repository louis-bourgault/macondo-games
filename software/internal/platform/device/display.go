//go:build device

// this will not compile with normal golang, since the machine library is only available in tinyGO
package device

import (
	"image/color"
	"machine"
)

type Display struct {
	spi *machine.SPI
	dc  machine.Pin
	cs  machine.Pin
	rst machine.Pin
	buf [240 * 240 * 2]byte
}

func NewDisplay(spi *machine.SPI, dcPin, csPin, rstPin machine.Pin) *Display {
	d := &Display{
		spi: spi,
		dc:  dcPin,
		cs:  csPin,
		rst: rstPin,
	}
	d.init()
	return d
}

func (d *Display) init() {
	d.rst.Low()
	d.rst.High()
}

func (d *Display) FillScreen(c color.Color) {
	r, g, b, _ := c.RGBA()
	pixel := rgb565((uint8)(r>>8), (uint8)(g>>8), (uint8)(b>>8))
	for i := 0; i < len(d.buf); i += 2 {
		d.buf[i] = (uint8)(pixel >> 8)
		d.buf[i+1] = (uint8)(pixel & 0xFF)
	}
	d.Present()
}

func (d *Display) DrawPixel(x, y int, c color.Color) {
	if x < 0 || x >= 240 || y < 0 || y >= 240 {
		return
	}
	r, g, b, _ := c.RGBA()
	pixel := rgb565((uint8)(r>>8), (uint8)(g>>8), (uint8)(b>>8))
	idx := (y*240 + x) * 2
	d.buf[idx] = (uint8)(pixel >> 8)
	d.buf[idx+1] = (uint8)(pixel & 0xFF)
}

func (d *Display) DrawRect(x, y, w, h int, c color.Color) {
	for i := 0; i < w; i++ {
		d.DrawPixel(x+i, y, c)
		d.DrawPixel(x+i, y+h-1, c)
	}
	for i := 0; i < h; i++ {
		d.DrawPixel(x, y+i, c)
		d.DrawPixel(x+w-1, y+i, c)
	}
}

func (d *Display) FillRect(x, y, w, h int, c color.Color) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			d.DrawPixel(xx, yy, c)
		}
	}
}

func (d *Display) DrawLine(x1, y1, x2, y2 int, c color.Color) {
	// Same Bresenham as desktop
	dx := x2 - x1
	dy := y2 - y1
	steps := dx
	if dy > dx {
		steps = dy
	}
	if steps == 0 {
		d.DrawPixel(x1, y1, c)
		return
	}
	for i := 0; i <= steps; i++ {
		x := x1 + (i*dx)/steps
		y := y1 + (i*dy)/steps
		d.DrawPixel(x, y, c)
	}
}

func (d *Display) Present() error {
	// Send framebuffer to LCD via SPI
	d.cs.Low()
	d.dc.High()
	d.spi.Tx(d.buf[:], nil)
	d.cs.High()
	return nil
}

func rgb565(r, g, b uint8) uint16 {
	return (uint16(r&0xF8) << 8) | (uint16(g&0xFC) << 3) | (uint16(b&0xF8) >> 3)
}
