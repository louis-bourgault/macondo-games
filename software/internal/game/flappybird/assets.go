package flappybird

import (
	_ "embed"

	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
)

//go:embed bird.bin
var rawBirdData string

var BirdSprite = helpers.Image{
	Data: rawBirdData,
	W:    12,
	H:    12,
}

//go:embed flappybackground.bin
var backgroundData string

var backgroundImage = helpers.Image{
	Data: backgroundData,
	W:    240,
	H:    240,
}
