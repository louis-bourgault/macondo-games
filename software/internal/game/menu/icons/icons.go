package icons

import (
	_ "embed"

	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
)

//go:embed snake.bin
var SnakeIcon string

var SnakeImage = helpers.Image{
	Data: SnakeIcon,
	W:    50,
	H:    50,
}
