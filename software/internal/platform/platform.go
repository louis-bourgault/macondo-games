package platform

type Screen interface {
	Fill(color uint16)
	Pixel(x, y int, c uint16)
	Present() error
}
