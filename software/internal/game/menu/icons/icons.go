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

//go:embed flappybird.bin
var FlappyBirdIcon string

var FlappyBirdImage = helpers.Image{
	Data: FlappyBirdIcon,
	W:    50,
	H:    50,
}

//go:embed mandelbrot.bin
var MandelbrotIcon string

var MandelbrotImage = helpers.Image{
	Data: MandelbrotIcon,
	W:    50,
	H:    50,
}

//go:embed julia.bin
var JuliaIcon string

var JuliaImage = helpers.Image{
	Data: JuliaIcon,
	W:    50,
	H:    50,
}

//go:embed chess.bin
var ChessIcon string

var ChessImage = helpers.Image{
	Data: ChessIcon,
	W:    50,
	H:    50,
}

//go:embed blackjack.bin
var BlackjackIcon string //we use the chess icon for now, TODO: fix

var BlackjackImage = helpers.Image{
	Data: BlackjackIcon,
	W:    50,
	H:    50,
}

//go:embed tetris.bin
var TetrisIcon string //we use the chess icon for now, TODO: fix

var TetrisImage = helpers.Image{
	Data: TetrisIcon,
	W:    50,
	H:    50,
}
