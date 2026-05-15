package game

import "github.com/louis-bourgault/macondo-games/software/internal/platform"

type Game interface {
	Update()
	Draw(screen platform.Screen)
}
