package tetris

import (
	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

// some of them have a 4x4 bounding box, and some have a 3x3.
// we can ignroe this, and just model them all as a 4x4 with the stuff in the top right if its not good
var tetronimoes = []tetronimo{
	{
		color: 0x037f,
		rotations: []uint16{
			3840, 8738, 240, 17476,
		},
	},
	{
		color: 0x8c71,
		rotations: []uint16{
			52224, 52224, 52224, 52224,
		},
	},
}

type tetronimo struct {
	rotations []uint16 //a byte for each level of rotion, since its a 4x4 grid
	color     uint16
}

type TetrisGame struct {
	blocks               [300]uint16 //the extra 10 at the top are to store the
	currentTopLeftX      int8        //the x position of the top left of the bounding box.
	currentTopLeftY      int8        //the y pos of the top left of the bounding box
	currentRotationIndex int8
	currentTetromino     tetronimo
}

func New() *TetrisGame {
	return &TetrisGame{
		blocks:               [300]uint16{},
		currentTopLeftX:      5,
		currentTopLeftY:      25,
		currentTetromino:     tetronimoes[helpers.RandomInt(0, len(tetronimoes)-1)],
		currentRotationIndex: 0,
	}
}

func (d *TetrisGame) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	return d
}

func (d *TetrisGame) Draw(screen platform.Screen) {
	screen.Fill(0x0000)
	screen.FillRect(60, 0, 120, 240, 0xd6ba)
	screen.Present() //present the screen
}
