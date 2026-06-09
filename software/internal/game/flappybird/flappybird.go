package flappybird

//a template game to copy and build off.

import (
	"strconv"

	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *FlappyBirdGame {
	return &FlappyBirdGame{
		birdY:        120,
		birdVelocity: 0,
		init:         false,
		gameover:     false,
		pipe1:        100,
		pipe2:        150,
		pipe1X:       240,
		pipe2X:       360,
		score:        0,
	}
}

type FlappyBirdGame struct {
	birdY        float32
	birdVelocity float32
	init         bool
	gameover     bool
	pipe1        uint8 //these are floats of the y position of the top of the gap in the pipes. The bottom of the gap is just 60 pixels below that.
	pipe2        uint8
	pipe1X       float32 //how far along the screen it is. When it reaches 0, it resets to 240 and the y position is randomized.
	pipe2X       float32
	score        int
}

const accel = 100
const clickboost = 80

func (d *FlappyBirdGame) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	if input.WasKeyJustPressed(platform.A) {
		log.Log("flap")
		d.birdVelocity = clickboost
		d.init = true
	}
	if d.gameover {
		if input.WasKeyJustPressed(platform.A) {
			return New()
		} else {
			return d
		}
	}
	if d.init {
		d.birdVelocity -= accel * float32(dt)
		d.birdY = (d.birdY) - (d.birdVelocity * float32(dt)) //the bird's y position is an unsigned int, but the velocity is a signed float, so we have to convert them to the same type before we can do the math. We also have to negate the velocity because in our coordinate system, up is negative and down is positive.
		if d.birdY < 0 {
			d.gameover = true
			d.birdY = 0
			log.Log("game over")
			return d
		}
		if d.birdY > 240-float32(BirdSprite.H) {
			d.gameover = true
			d.birdY = 240 - float32(BirdSprite.H)
			log.Log("game over")
			return d
		}
		d.pipe1X -= 60 * float32(dt)
		d.pipe2X -= 60 * float32(dt)
		if d.pipe1X < -20 {
			d.pipe1X = 240
			d.pipe1 = uint8(helpers.RandomInt(30, 170))
		}
		if d.pipe2X < -20 {
			d.pipe2X = 240
			d.pipe2 = uint8(helpers.RandomInt(30, 170))
		}
		if (d.pipe1X < float32(80+BirdSprite.W) && d.pipe1X > float32(80-BirdSprite.W)) && (d.birdY < float32(d.pipe1) || d.birdY > float32(d.pipe1+60)) {
			d.gameover = true
			log.Log("game over")
			return d
		}
		if (d.pipe2X < float32(80+BirdSprite.W) && d.pipe2X > float32(80-BirdSprite.W)) && (d.birdY < float32(d.pipe2) || d.birdY > float32(d.pipe2+60)) {
			d.gameover = true
			log.Log("game over")
			return d
		}
		//if we pass by a pipe, add to the score
		if d.pipe1X < 80 && d.pipe1X > 80-60*float32(dt) {
			d.score++
		}
		if d.pipe2X < 80 && d.pipe2X > 80-60*float32(dt) {
			d.score++
		}

	}
	return d
}

func (d *FlappyBirdGame) Draw(screen platform.Screen) {
	if d.gameover {
		screen.Fill(0x0000)
		helpers.DrawText(screen, 50, 100, "game over", 0xF800)
		helpers.DrawText(screen, 20, 120, "press A to restart", 0xF800)
		helpers.DrawText(screen, 20, 140, "score: "+strconv.Itoa(d.score), 0x001f)
		screen.Present()
		return
	} else {
		helpers.DrawImage(screen, backgroundImage, 0, 0)
		helpers.DrawImage(screen, BirdSprite, 80, int16(d.birdY))
		screen.FillRect(int(d.pipe1X), 0, 20, int(d.pipe1), 0x07E0)
		screen.FillRect(int(d.pipe1X), int(d.pipe1+60), 20, int(240-d.pipe1-60), 0x07E0)
		screen.FillRect(int(d.pipe2X), 0, 20, int(d.pipe2), 0x07E0)
		screen.FillRect(int(d.pipe2X), int(d.pipe2+60), 20, int(240-d.pipe2-60), 0x07E0)
		helpers.DrawText(screen, 5, 5, "score: "+strconv.Itoa(d.score), 0xF800)
		screen.Present() //present the screen

	}
}
