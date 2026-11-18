package main

import (
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
	"github.com/louis-bourgault/macondo-games/software/internal/platform/device"
)

func newGPIOSystem() platform.GPIOSystem {
	return device.NewGPIOSystem()
}
