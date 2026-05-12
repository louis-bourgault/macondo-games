package platform

import "image/color"

type Display interface {
	FillScreen(c color.Color)
	DrawPixel(x, y int, c color.Color)
	DrawRect(x, y, width, height int, c color.Color, fill bool)
	Present() error
	//could add circles and stuff
}

type Input interface {
	Update()
	Pressed(key string) bool
	JustPressed(key string) bool
}

type Clock interface {
	Sleep(ms int)
	Delta() float64
	FPS() float64
}

const (
	BtnUp     = "up"
	BtnDown   = "down"
	BtnLeft   = "left"
	BtnRight  = "right"
	BtnA      = "a"
	BtnB      = "b"
	BtnStart  = "start"
	BtnSelect = "select"
)
