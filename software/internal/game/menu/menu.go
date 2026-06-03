package menu

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/game/drawtest"
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
		}, position: 0,
	}
}

type Menu struct {
	gameOptions []struct {
		name     string
		launcher func() game.Game
	}
	position int8
}

func (m *Menu) Update(dt float64, input platform.InputSystem) game.Game {
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
	for i, option := range m.gameOptions {
		if i == int(m.position) {
			helpers.DrawText(screen, 2, (i*8 + 2), option.name, 0x001f)
		} else {
			helpers.DrawText(screen, 2, (i*8 + 2), option.name, 0x0000)
		}
	}
	screen.Present()
}
