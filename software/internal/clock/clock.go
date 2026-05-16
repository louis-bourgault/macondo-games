package clock

import "time"

type Clock struct {
	last  time.Time
	delta float64
	fps   float64
}

func New() *Clock {
	return &Clock{
		last:  time.Now(),
		delta: 0,
		fps:   0,
	}
}

func (c *Clock) Tick() {
	now := time.Now()
	dt := now.Sub(c.last).Seconds()
	if dt <= 0 {
		dt = 1.0 / 60.0
	}
	c.delta = dt
	c.fps = 1.0 / dt
	c.last = now
}

func (c *Clock) Delta() float64 { return c.delta }
func (c *Clock) FPS() float64   { return c.fps }
func (c *Clock) Sleep(ms int)   { time.Sleep(time.Duration(ms) * time.Millisecond) }
