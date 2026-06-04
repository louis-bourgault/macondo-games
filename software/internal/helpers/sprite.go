package helpers

import (
	"unsafe"

	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

type Sprite struct {
	Data string
	W    uint8
	H    uint8
}

func (s Sprite) GetPixel(row, col uint8) uint16 {
	index := (int(row) * int(s.W)) + int(col)
	//pointer to the start of the data
	strPtr := unsafe.StringData(s.Data)

	// Cast the byte pointer to a uint16 pointer and offset it by our index.
	//this lets us read from rom, not having to pass stuff through our limited ram
	pixelPtr := (*uint16)(unsafe.Pointer(strPtr))

	return *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(pixelPtr)) + uintptr(index)*2))
}

func DrawSprite(screen platform.Screen, sprite Sprite, x, y int8) {
	var row uint8
	var col uint8

	for row = 0; row < sprite.H; row++ {
		currentY := int(y) + int(row)
		if currentY < 0 || currentY >= 240 { //check if its out of bounds
			continue
		}

		for col = 0; col < sprite.W; col++ {
			currentX := int(x) + int(col)
			if currentX < 0 || currentX >= 240 {
				continue
			}
			//can grab it from flash and send it straight to the screen
			color := sprite.GetPixel(row, col)
			screen.Pixel(currentX, currentY, color)
		}
	}
}
