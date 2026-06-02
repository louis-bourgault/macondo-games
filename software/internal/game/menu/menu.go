package menu

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/game/drawtest"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *Menu {
	return &Menu{}
}

type Menu struct {
}

func (m *Menu) Update(dt float64, input platform.InputSystem) game.Game {
	if input.WasKeyJustPressed("A") {
		return drawtest.New()
	}
	return m //if nothing is selected, this just returns the game we are currently on, the menu
}

func (m *Menu) Draw(screen platform.Screen) {
	screen.Fill(0x07e0)
	screen.Present()
}
