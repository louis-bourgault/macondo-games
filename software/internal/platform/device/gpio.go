package device

type GPIOSystem struct {
	enabled bool
}

func NewGPIOSystem() *GPIOSystem {
	if !gpioAvailable {
		return &GPIOSystem{
			enabled: false,
		}
	}

}

//all the pins on the v2 system are analog or digital capable

func (g *GPIOSystem) Enabled() bool {
	return g.enabled
}

func (g *GPIOSystem) Enable() {
	//by default its enabled
	g.enabled = true
}

func (g *GPIOSystem) Disable() {
	//by default its enabled, but you can turn it off
	g.enabled = false
}

func (g *GPIOSystem) AnalogRead(pin int) int {
	if !g.enabled {
		return 0
	}

	if pin < 1 || pin > len(gpioPins) {
		return 0
	}

	return int(gpioPins[pin-1].Pin.Get())
}

func (g *GPIOSystem) DigitalRead(pin int) bool {
	if !g.enabled {
		return false
	}

	if pin < 1 || pin > len(gpioPins) {
		return false
	}

	return gpioPins[pin-1].Pin.Get() != 0
}

func (g *GPIOSystem) DigitalWrite(pin int, value bool) {
	if !g.enabled {
		return
	}

	if pin < 1 || pin > len(gpioPins) {
		return
	}

	if value {
		gpioPins[pin-1].Pin.High()
	} else {
		gpioPins[pin-1].Pin.Low()
	}
}

func (g *GPIOSystem) AnalogWrite(pin int, value int) {
	if !g.enabled {
		return
	}

	if pin < 1 || pin > len(gpioPins) {
		return
	}

	gpioPins[pin-1].Pin.Set(value)
}
