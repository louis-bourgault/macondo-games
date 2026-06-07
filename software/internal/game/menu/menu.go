package menu

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/game/chess"
	"github.com/louis-bourgault/macondo-games/software/internal/game/drawtest"
	"github.com/louis-bourgault/macondo-games/software/internal/game/flappybird"
	"github.com/louis-bourgault/macondo-games/software/internal/game/julia"
	"github.com/louis-bourgault/macondo-games/software/internal/game/mandelbrot"
	"github.com/louis-bourgault/macondo-games/software/internal/game/snake"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *Menu {
	return &Menu{
		gameOptions: []struct {
			name     string
			launcher func() game.Game
		}{
			{name: "drawtest", launcher: func() game.Game { return drawtest.New() }},
			{name: "snake", launcher: func() game.Game { return snake.New() }},
			{name: "flappy bird", launcher: func() game.Game { return flappybird.New() }},
			{name: "mandelbrot", launcher: func() game.Game { return mandelbrot.New() }},
			{name: "julia", launcher: func() game.Game { return julia.New() }},
			{name: "chess", launcher: func() game.Game { return chess.New() }},
		}, position: 0,
	}
}

type Menu struct {
	gameOptions []struct {
		name     string
		launcher func() game.Game
		//icon helpers.Image
	}
	position int8
}

func (m *Menu) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Down) {
		if int(m.position) < len(m.gameOptions)-1 {
			m.position++
		}
	}
	if input.WasKeyJustPressed(platform.Up) {
		if m.position > 0 {
			m.position--
		}
	}
	if input.WasKeyJustPressed(platform.A) {
		return m.gameOptions[m.position].launcher()
	}
	return m //if nothing is selected, this just returns the game we are currently on, the menu
}

func (m *Menu) Draw(screen platform.Screen) {
	screen.Fill(0xffff)
	//each one is 50 high and 50 wide, and we show 3 of them

	// helpers.DrawImage(screen, i  icons.SnakeImage, 15, 15)

	for i, option := range m.gameOptions {
		// if i < int(m.position-1) {
		// 	continue
		// }
		if i == int(m.position) {
			helpers.DrawText(screen, 2, (i*8 + 2), option.name, 0x001f)
		} else {
			helpers.DrawText(screen, 2, (i*8 + 2), option.name, 0x0000)
		}
	}
	screen.Present()
}
