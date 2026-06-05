package platform

type Screen interface {
	Fill(color uint16)
	Pixel(x, y int, c uint16)
	FillRect(x, y, w, h int, c uint16)
	Present() error
} //could change some of these data types from int to uint8 to optimise a few more bytes away, but its pretty insignificant

type InputSystem interface {
	WasKeyJustPressed(key Button) bool
	WasKeyJustReleased(key Button) bool
	IsKeyPressed(key Button) bool
	Update()
}

type LogSystem interface {
	Log(log string)
}

type Button uint8

const (
	None Button = iota //make sure that none is equal to 0
	Up
	Down
	Left
	Right
	A
	B
	Start
	Select
)
