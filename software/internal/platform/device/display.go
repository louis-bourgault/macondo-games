package device

import (
	"machine"
	"time"

	"unsafe"

	"tinygo.org/x/drivers/st7789"
)

//TODO: i don't know how performant this is. Would it be faster to use a buffer then push it to spi?

const (
	//SPI0 SCLK - Pin 18
	PIN_SPI_SCK = machine.SPI0_SCK_PIN
	//SPI0 TX/MOSI - Pin 19
	PIN_SPI_SDO = machine.SPI0_SDO_PIN
	PIN_DC      = machine.Pin(21)
	PIN_RST     = machine.Pin(20)
	CS_SCREEN   = machine.Pin(22)
)

type HardwareDisplay struct {
	dev *st7789.Device
	spi *machine.SPI
	dc  machine.Pin
	//this takes up 115200 bytes, or 115kb of the total 264kb of ram on the device.
	//it is possible to make it bufferless and write direct to the display, but performance would be a lot worse, and i think 149kb is probably enough for everything else.
	buf [240 * 240]uint16
}

func (d *HardwareDisplay) Fill(c uint16) {
	for i := range d.buf {
		d.buf[i] = c
	}
}

func (d *HardwareDisplay) FillRect(x, y, w, h int, c uint16) {
	for sy := y; sy < y+h; sy++ {
		row := sy * 240
		for sx := x; sx < x+w; sx++ {
			d.buf[row+sx] = c
		}
	}
}

func (d *HardwareDisplay) Pixel(x, y int, c uint16) {
	if x < 0 || x >= 240 || y < 0 || y >= 240 {
		return
	}
	d.buf[y*240+x] = c
}

func (d *HardwareDisplay) Present() error {
	//now we use a buffer, we actually need a present function
	//might need a set window thing here
	d.dc.High()
	byteData := unsafe.Slice((*byte)(unsafe.Pointer(&d.buf[0])), len(d.buf)*2)
	d.spi.Tx(byteData, nil)
	return nil
}

func NewDisplay() *HardwareDisplay {
	CS_SCREEN.Low()

	spi := machine.SPI0
	spi.Configure(machine.SPIConfig{
		Frequency: 8000000,
		SCK:       PIN_SPI_SCK,
		SDO:       PIN_SPI_SDO,
	})
	PIN_DC.Configure(machine.PinConfig{Mode: machine.PinOutput})
	PIN_RST.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dev := st7789.New(spi, PIN_DC, PIN_RST, CS_SCREEN, machine.NoPin)
	dev.Configure(st7789.Config{
		Width:  240,
		Height: 240,
	})
	time.Sleep(100 * time.Millisecond)
	return &HardwareDisplay{
		dev: &dev,
		spi: spi,
		dc:  PIN_DC,
	}
}
