package device

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/st7789"
)

//TODO: i don't know how performant this is. Would it be faster to use a buffer then push it to spi?

const (
	//SPI0 SCLK - Pin 18
	PIN_SPI_SCK = machine.SPI0_SCK_PIN
	//SPI0 TX/MOSI - Pin 19
	PIN_SPI_SDO = machine.SPI0_SDO_PIN
	PIN_DC      = machine.PIN(21)
	PIN_RST     = machine.PIN(20)
	CS_SCREEN   = machine.GPIO{machine.PIN(22)}
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
	//st7789 updates immediately
	return nil
}

func NewDisplay() *st7789.Device {
	CS_SCREEN.Low()

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
