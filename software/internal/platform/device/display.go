package device

//go complains, but works with tinygo.
import (
	"machine"
	"time"

	"unsafe"

	"tinygo.org/x/drivers/st7789"
)

const ChromaKey uint16 = 0xF81F

const (
	//SPI0 SCLK - Pin 18
	PIN_SPI_SCK = machine.Pin(18)
	//SPI0 TX/MOSI - Pin 19
	PIN_SPI_SDO = machine.Pin(19)
	PIN_DC      = machine.Pin(21)
	PIN_RST     = machine.Pin(20)
	CS_SCREEN   = machine.Pin(17)
)

type HardwareDisplay struct {
	dev *st7789.Device
	spi *machine.SPI
	dc  machine.Pin
	cs  machine.Pin
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
		dy := 239 - sy
		if dy < 0 || dy >= 240 {
			continue
		}
		row := dy * 240
		for sx := x; sx < x+w; sx++ {
			dx := 239 - sx
			if dx < 0 || dx >= 240 {
				continue
			}
			d.buf[row+dx] = c
		}
	}
}

func (d *HardwareDisplay) Pixel(x, y int, c uint16) {
	if c == ChromaKey {
		return
	}
	if x < 0 || x >= 240 || y < 0 || y >= 240 {
		return
	}
	d.buf[(239-y)*240+(239-x)] = c
}

func (d *HardwareDisplay) Present() error {

	d.cs.Low()

	//set the window to the full display area with 0x2A (column address set)
	d.dc.Low()
	d.spi.Tx([]byte{0x2A}, nil)
	d.dc.High()
	d.spi.Tx([]byte{0x00, 0x00, 0x00, 0xEF}, nil) //0-239

	//send command 0x2B(row address set)
	d.dc.Low()
	d.spi.Tx([]byte{0x2B}, nil)
	d.dc.High()
	d.spi.Tx([]byte{0x00, 0x00, 0x00, 0xEF}, nil) //row0-239

	//send command 0x2C(write memory)
	d.dc.Low()
	d.spi.Tx([]byte{0x2C}, nil)
	d.dc.High()

	//make bigendian for st7789 chip
	for i := range d.buf {
		d.buf[i] = (d.buf[i] << 8) | (d.buf[i] >> 8)
	}

	byteData := unsafe.Slice((*byte)(unsafe.Pointer(&d.buf[0])), len(d.buf)*2)
	d.spi.Tx(byteData, nil)

	//swap back to littleendian just in case. mihgt remove later
	for i := range d.buf {
		d.buf[i] = (d.buf[i] << 8) | (d.buf[i] >> 8)
	}

	d.cs.High()
	return nil
}

func NewDisplay() *HardwareDisplay {
	PIN_DC.Configure(machine.PinConfig{Mode: machine.PinOutput})
	PIN_RST.Configure(machine.PinConfig{Mode: machine.PinOutput})
	CS_SCREEN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	CS_SCREEN.High()

	spi := machine.SPI0
	spi.Configure(machine.SPIConfig{
		Frequency: 20_000_000,
		SCK:       PIN_SPI_SCK,
		SDO:       PIN_SPI_SDO,
		SDI:       machine.Pin(16), //not used, but there's internal st7789 driver validation that this is a valid pin.
		Mode:      0,
	})

	PIN_RST.High()
	time.Sleep(10 * time.Millisecond)
	PIN_RST.Low()
	time.Sleep(10 * time.Millisecond)
	PIN_RST.High()
	time.Sleep(120 * time.Millisecond)

	dev := st7789.New(spi, PIN_RST, PIN_DC, CS_SCREEN, machine.NoPin)
	dev.Configure(st7789.Config{
		Width:  240,
		Height: 240,
	})

	println("Display initialized")
	return &HardwareDisplay{
		dev: &dev,
		spi: spi,
		dc:  PIN_DC,
		cs:  CS_SCREEN,
	}
}
