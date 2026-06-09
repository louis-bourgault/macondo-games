package menu

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/game/blackjack"
	"github.com/louis-bourgault/macondo-games/software/internal/game/chess"
	"github.com/louis-bourgault/macondo-games/software/internal/game/flappybird"
	"github.com/louis-bourgault/macondo-games/software/internal/game/julia"
	"github.com/louis-bourgault/macondo-games/software/internal/game/mandelbrot"
	"github.com/louis-bourgault/macondo-games/software/internal/game/menu/icons"
	"github.com/louis-bourgault/macondo-games/software/internal/game/snake"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *Menu {
	return &Menu{
		gameOptions: []struct {
			name     string
			launcher func() game.Game
			icon     helpers.Image
		}{
			// {name: "drawtest", launcher: func() game.Game { return drawtest.New() }},
			{name: "Snake", launcher: func() game.Game { return snake.New() }, icon: icons.SnakeImage},
			{name: "Flappy Bird", launcher: func() game.Game { return flappybird.New() }, icon: icons.FlappyBirdImage},
			{name: "Mandelbrot Set", launcher: func() game.Game { return mandelbrot.New() }, icon: icons.MandelbrotImage},
			{name: "Julia Set", launcher: func() game.Game { return julia.New() }, icon: icons.JuliaImage},
			{name: "Chess", launcher: func() game.Game { return chess.New() }, icon: icons.ChessImage},
			{name: "Blackjack", launcher: func() game.Game { return blackjack.New() }, icon: icons.BlackjackImage},
		}, position: 0,
	}
}

type Menu struct {
	gameOptions []struct {
		name     string
		launcher func() game.Game
		icon     helpers.Image
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
	screen.Fill(0x0000)
	//each one is 50 high and 50 wide, and we show 3 of them

	if m.position > 0 {
		helpers.DrawImage(screen, m.gameOptions[m.position-1].icon, 30, 30)
		helpers.DrawText(screen, 95, 59, m.gameOptions[m.position-1].name, 0xffff)
		helpers.DrawImage(screen, m.gameOptions[m.position].icon, 30, 95)
		helpers.DrawText(screen, 95, 124, m.gameOptions[m.position].name, 0xF800)
		if int(m.position) < len(m.gameOptions)-1 {
			helpers.DrawImage(screen, m.gameOptions[m.position+1].icon, 30, 160)
			helpers.DrawText(screen, 95, 189, m.gameOptions[m.position+1].name, 0xffff)
		}
	} else {
		helpers.DrawImage(screen, m.gameOptions[m.position].icon, 30, 30)
		helpers.DrawText(screen, 95, 59, m.gameOptions[m.position].name, 0xF800)
		if int(m.position) < len(m.gameOptions)-1 {
			helpers.DrawImage(screen, m.gameOptions[m.position+1].icon, 30, 95)
			helpers.DrawText(screen, 95, 124, m.gameOptions[m.position+1].name, 0xffff)
			if int(m.position) < len(m.gameOptions)-2 {
				helpers.DrawImage(screen, m.gameOptions[m.position+2].icon, 30, 160)
				helpers.DrawText(screen, 95, 189, m.gameOptions[m.position+2].name, 0xffff)
			}
		}
	}
	helpers.DrawText(screen, 76, 2, "FERRETBOARD", 0xffff)
	helpers.DrawText(screen, 52, 230, "PRESS A TO SELECT", 0xffff)

	// for i, option := range m.gameOptions {
	// 	// if i < int(m.position-1) {
	// 	// 	continue
	// 	// }
	// 	if i == int(m.position) {
	// 		helpers.DrawText(screen, 2, (i*8 + 2), option.name, 0x001f)
	// 	} else {
	// 		helpers.DrawText(screen, 2, (i*8 + 2), option.name, 0x0000)
	// 	}
	// }
	screen.Present()
}
