package game

import "github.com/louis-bourgault/macondo-games/software/internal/platform"

type Game interface {
	Update(dt float64)
	Draw(screen platform.Screen)
}
