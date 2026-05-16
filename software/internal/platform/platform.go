package platform

type Screen interface {
	Fill(color uint16)
	Pixel(x, y int, c uint16)
	Present() error
}

type InputSystem interface {
	WasKeyJustPressed(key string) bool
	WasKeyJustReleased(key string) bool
	IsKeyPressed(key string) bool
	Update()
}
