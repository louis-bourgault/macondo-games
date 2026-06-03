package platform

type Screen interface {
	Fill(color uint16)
	Pixel(x, y int, c uint16)
	FillRect(x, y, w, h int, c uint16)
	Present() error
}

type InputSystem interface {
	WasKeyJustPressed(key Button) bool
	WasKeyJustReleased(key Button) bool
	IsKeyPressed(key Button) bool
	Update()
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
