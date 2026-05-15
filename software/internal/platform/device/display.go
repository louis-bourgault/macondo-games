package device

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/st7789"
)

const (
	PIN_SPI_SCK = machine.SPI0_SCK_PIN
	PIN_SPI_SDO = machine.SPI0_SDO_PIN
	PIN_DC      = machine.PIN(8)
	PIN_RST     = machine.PIN(9)
)

type HardwareDisplay struct {
	dev *st7789.Device
}

func (d *HardwareDisplay) Fill(c uint16) {
	col := color.RGBA{
		R: uint8((c >> 11) & 0x1f << 3),
		G: uint8((c >> 5) & 0x3f << 2),
		B: uint8((c & 0x1f) << 3),
		A: 0xff,
	}
	d.dev.FillScreen(col)
}

func (d *HardwareDisplay) Pixel(x, y int, c uint16) {
	col := color.RGBA{
		R: uint8((c >> 11) & 0x1f << 3),
		G: uint8((c >> 5) & 0x3f << 2),
		B: uint8((c & 0x1f) << 3),
		A: 0xff,
	}
	d.dev.SetPixel(x, y, col)
}

func (d *HardwareDisplay) Present() error {
	//st7789 updates immediately, so don't need to do anything here
	return nil
}

func NewDisplay() *st7789.Device {
	spi := machine.SPI0
	spi.Configure(machine.SPIConfig{
		Frequency: 8000000,
		SCK:       PIN_SPI_SCK,
		SDO:       PIN_SPI_SDO,
	})
	dc := machine.GPIO{PIN_DC}
	rst := machine.GPIO{PIN_RST}
	dc.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	rst.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	dev := st7789.New(spi, dc, rst)
	dev.Configure(st7789.Config{
		Width:  240,
		Height: 240,
	})
	time.Sleep(100 * time.Millisecond)
	return dev
}
