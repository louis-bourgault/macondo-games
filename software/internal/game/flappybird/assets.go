package flappybird

import (
	_ "embed"

	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
)

//go:embed bird.bin
var rawBirdData string

var BirdSprite = helpers.Sprite{
	Data: rawBirdData,
	W:    12,
	H:    12,
}
