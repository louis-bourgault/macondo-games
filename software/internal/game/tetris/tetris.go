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
		color: 0x037f, //long boi
		rotations: []uint16{
			3840, 8738, 240, 17476,
		},
	},
	{
		color: 0x8c71, //square
		rotations: []uint16{
			52224, 52224, 52224, 52224,
		},
	},
	{
		color: 0xfcb6, //cant remember what the rest of them are
		rotations: []uint16{
			36352, 25664, 3616, 17600,
		},
	},
	{
		color: 0x47f4,
		rotations: []uint16{
			11776, 17504, 3712, 50240,
		},
	},
	{
		color: 0x0621,
		rotations: []uint16{
			27648, 17952, 1728, 35904,
		},
	},
	{
		color: 0x981c,
		rotations: []uint16{
			19968, 17984, 3648, 19520,
		},
	},
	{
		color: 0xe380,
		rotations: []uint16{
			50688, 9792, 3168, 19584,
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
	gameOver             bool
	timeSince            float64
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
	if input.WasKeyJustPressed(platform.Select) {
		return nil
	}
	if d.gameOver {
		if input.WasKeyJustPressed(platform.A) {
			return New()
		}
		return d
	}
	if input.WasKeyJustPressed(platform.Left) {
		d.moveLeft()
	}
	if input.WasKeyJustPressed(platform.Right) {
		d.moveRight()
	}
	if input.WasKeyJustPressed(platform.Down) {
		for !d.gameOver {
			oldY := d.currentTopLeftY
			d.moveDown()
			if d.currentTopLeftY == 25 || d.currentTopLeftY == oldY {
				break
			}
		}
	}
	if input.WasKeyJustPressed(platform.Up) {
		d.rotate()
	}
	d.timeSince += dt
	if d.timeSince > 0.5 {
		d.moveDown()
		d.timeSince = 0
	}
	return d
}

type location struct {
	x int8
	y int8
}

func getNthBit(int uint16, n uint8) bool {
	return (int & (1 << n)) != 0
}

func (d *TetrisGame) moveLeft() {
	var occupiedPositions []location
	copy := d.currentTetromino.rotations[d.currentRotationIndex]

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if getNthBit(copy, uint8(i*4+j)) {
				occupiedPositions = append(occupiedPositions, location{
					x: d.currentTopLeftX + int8(i),
					y: d.currentTopLeftY + int8(j),
				})
			}
		}
	}
	for _, pos := range occupiedPositions {
		if pos.x <= 0 || pos.y < 10 || pos.y >= 30 {
			return
		}
		if d.blocks[int(pos.y)*10+int(pos.x)-1] != 0 {
			return
		}
	}
	d.currentTopLeftX -= 1
}

func (d *TetrisGame) moveRight() {
	var occupiedPositions []location
	copy := d.currentTetromino.rotations[d.currentRotationIndex]

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if getNthBit(copy, uint8(i*4+j)) {
				occupiedPositions = append(occupiedPositions, location{
					x: d.currentTopLeftX + int8(i),
					y: d.currentTopLeftY + int8(j),
				})
			}
		}
	}
	for _, pos := range occupiedPositions {
		if pos.x >= 9 || pos.y < 10 || pos.y >= 30 {
			return
		}
		if d.blocks[int(pos.y)*10+int(pos.x)+1] != 0 {
			return
		}
	}
	d.currentTopLeftX += 1
}

func (d *TetrisGame) moveDown() {
	var occupiedPositions []location
	copy := d.currentTetromino.rotations[d.currentRotationIndex]

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if getNthBit(copy, uint8(i*4+j)) {
				occupiedPositions = append(occupiedPositions, location{
					x: d.currentTopLeftX + int8(i),
					y: d.currentTopLeftY + int8(j),
				})
			}
		}
	}
	for _, pos := range occupiedPositions {
		if pos.y <= 10 {
			d.lockTetromino(occupiedPositions)
			d.getNewTetromino()
			return
		}
		index := int(pos.y-1)*10 + int(pos.x)
		if index >= 0 && index < 300 && d.blocks[index] != 0 {
			d.lockTetromino(occupiedPositions)
			d.getNewTetromino()
			return
		}
	}
	d.currentTopLeftY -= 1
}

func (d *TetrisGame) rotate() {
	var occupiedPositions []location
	copy := d.currentTetromino.rotations[(d.currentRotationIndex+1)%int8(len(d.currentTetromino.rotations))]

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if getNthBit(copy, uint8(i*4+j)) {
				occupiedPositions = append(occupiedPositions, location{
					x: d.currentTopLeftX + int8(i),
					y: d.currentTopLeftY + int8(j),
				})
			}
		}
	}
	for _, pos := range occupiedPositions {
		if pos.x < 0 || pos.x >= 10 || pos.y < 10 || pos.y >= 30 {
			return
		}
		if d.blocks[int(pos.y)*10+int(pos.x)] != 0 {
			return
		}
	}
	d.currentRotationIndex = (d.currentRotationIndex + 1) % int8(len(d.currentTetromino.rotations))
}

func (d *TetrisGame) lockTetromino(occupiedPositions []location) {
	for _, pos := range occupiedPositions {
		if pos.x >= 0 && pos.x < 10 && pos.y >= 10 && pos.y < 30 {
			d.blocks[int(pos.y)*10+int(pos.x)] = d.currentTetromino.color
		}
	}
	d.clearLines()
}

func (d *TetrisGame) clearLines() {
	for y := 10; y < 30; y++ {
		full := true
		for x := 0; x < 10; x++ {
			if d.blocks[y*10+x] == 0 {
				full = false
				break
			}
		}
		if full {
			for ny := y; ny < 29; ny++ {
				for x := 0; x < 10; x++ {
					d.blocks[ny*10+x] = d.blocks[(ny+1)*10+x]
				}
			}
			for x := 0; x < 10; x++ {
				d.blocks[29*10+x] = 0
			}
			y--
		}
	}
}

func (d *TetrisGame) getNewTetromino() {
	d.currentTetromino = tetronimoes[helpers.RandomInt(0, len(tetronimoes)-1)]
	d.currentRotationIndex = 0
	d.currentTopLeftX = 5
	d.currentTopLeftY = 25
	//if it instantly collides, then its game over
	copy := d.currentTetromino.rotations[d.currentRotationIndex]
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if getNthBit(copy, uint8(i*4+j)) {
				x := d.currentTopLeftX + int8(i)
				y := d.currentTopLeftY + int8(j)
				if x >= 0 && x < 10 && y >= 10 && y < 30 {
					if d.blocks[int(y)*10+int(x)] != 0 {
						d.gameOver = true
						return
					}
				}
			}
		}
	}
}

func (d *TetrisGame) Draw(screen platform.Screen) {
	screen.Fill(0x0000)
	screen.FillRect(60, 0, 120, 240, 0xd6ba)

	if d.gameOver {
		helpers.DrawText(screen, 10, 10, "game over", 0xf800)
		helpers.DrawText(screen, 10, 50, "press A to restart", 0xf800)
		screen.Present()
		return
	}

	for y := 10; y < 30; y++ {
		for x := 0; x < 10; x++ {
			color := d.blocks[y*10+x]
			if color != 0 {
				screen.FillRect(x*12+60, (29-y)*12, 12, 12, color)
			}
		}
	}

	var occupiedPositions []location
	copy := d.currentTetromino.rotations[d.currentRotationIndex]

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if getNthBit(copy, uint8(i*4+j)) {
				occupiedPositions = append(occupiedPositions, location{
					x: d.currentTopLeftX + int8(i),
					y: d.currentTopLeftY + int8(j),
				})
			}
		}
	}
	for _, pos := range occupiedPositions {
		if pos.x >= 0 && pos.x < 10 && pos.y >= 10 && pos.y < 30 {
			screen.FillRect(int(pos.x)*12+60, (29-int(pos.y))*12, 12, 12, d.currentTetromino.color)
		}
	}

	screen.Present()
}
