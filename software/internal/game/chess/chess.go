package chess

//disclosure: substantial ai code, higher than the rest of this codebase. I still undertsand what everything does, was just getting tedious.
//all the decisions about layout of the code and what algorithms to use were made by me, and i know how they all work.

import (
	"fmt"

	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *ChessGame {
	return &ChessGame{
		board: [8][8]uint8{
			{0xC, 0xA, 0xB, 0xD, 0xE, 0xB, 0xA, 0xC}, // black: r n b q k b n r
			{0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9}, // black pawns
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}, // white pawns
			{0x4, 0x2, 0x3, 0x5, 0x6, 0x3, 0x2, 0x4}, // white: r n b q k b n r
		},
		selectedX:  0,
		selectedY:  0,
		moving:     false,
		playerTurn: true,
	}
}

type ChessGame struct {
	board                          [8][8]uint8
	selectedX                      int8
	selectedY                      int8
	moving                         bool
	selectedPiece                  uint8
	selectedPieceOriginalLocationX int8
	selectedPieceOriginalLocationY int8
	playerTurn                     bool //false = black, true = white
	//the player always plays white against the bot.
	//bit 0-2: piece type (0 = empty, 1 = pawn, 2 = knight, 3 = bishop, 4 = rook, 5 = queen, 6 = king)
	//bit 3: color (0 = white, 1 = black)
}

func (d *ChessGame) makeMove(move [4]int8) { //fromX, fromY, toX, toY
	piece := d.board[move[1]][move[0]]
	d.board[move[3]][move[2]] = piece
	d.board[move[1]][move[0]] = 0
}

func (d *ChessGame) unmakeMove(move [4]int8) {
	piece := d.board[move[3]][move[2]]
	d.board[move[1]][move[0]] = piece
	d.board[move[3]][move[2]] = 0
}

func (d *ChessGame) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) {
		return nil
	}
	if input.WasKeyJustPressed(platform.Down) {
		if d.selectedY < 7 {
			d.selectedY++
		}
	}
	if input.WasKeyJustPressed(platform.Up) {
		if d.selectedY > 0 {
			d.selectedY--
		}
	}
	if input.WasKeyJustPressed(platform.Left) {
		if d.selectedX > 0 {
			d.selectedX--
		}
	}
	if input.WasKeyJustPressed(platform.Right) {
		if d.selectedX < 7 {
			d.selectedX++
		}
	}
	if input.WasKeyJustPressed(platform.A) {
		if !d.moving {
			d.selectedPiece = d.board[d.selectedY][d.selectedX]
			if d.selectedPiece != 0 {
				d.moving = true
				d.selectedPieceOriginalLocationX = d.selectedX
				d.selectedPieceOriginalLocationY = d.selectedY
			}
		} else {
			moves := d.GenAllLegalMoves(d.playerTurn)
			isLegal := false
			for i := 0; i < moves.count; i++ {
				m := moves.moves[i]
				if m[0] == d.selectedPieceOriginalLocationX && m[1] == d.selectedPieceOriginalLocationY &&
					m[2] == d.selectedX && m[3] == d.selectedY {
					isLegal = true
					break
				}
			}
			if isLegal {
				// d.board[d.selectedY][d.selectedX] = d.selectedPiece
				// d.board[d.selectedPieceOriginalLocationY][d.selectedPieceOriginalLocationX] = 0
				d.makeMove([4]int8{d.selectedPieceOriginalLocationX, d.selectedPieceOriginalLocationY, d.selectedX, d.selectedY})
				d.playerTurn = !d.playerTurn
			} else {
				log.Log("illegal move")
			}
			d.moving = false
		}
	}
	if input.WasKeyJustPressed(platform.B) {
		log.Log("legal moves:" + fmt.Sprintf("%+v", d.GenAllLegalMoves(d.playerTurn)))
	}
	if !d.playerTurn {
		log.Log("Prior board state: " + fmt.Sprintf("%+v", d.board))
		var dCopy ChessGame
		dCopy.board = d.board
		dCopy.playerTurn = d.playerTurn
		bestMove, _ := chooseBestMove(&dCopy, 3)
		log.Log("best move: " + fmt.Sprintf("%+v", bestMove))
		d.makeMove(bestMove)
		log.Log("New board state: " + fmt.Sprintf("%+v", d.board))
		d.playerTurn = !d.playerTurn
	}
	return d
}

func (d *ChessGame) Draw(screen platform.Screen) {
	screen.Fill(0x3186) // dark grey border

	lightSquare := uint16(0xF6D9) // warm cream
	darkSquare := uint16(0x8A69)  // brown

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			posX := x*20 + 40
			posY := y*20 + 40

			// draw board square
			var sqColor uint16
			if (x+y)%2 == 0 {
				sqColor = lightSquare
			} else {
				sqColor = darkSquare
			}
			screen.FillRect(posX, posY, 20, 20, sqColor)

			// highlight selected square
			if int8(x) == d.selectedX && int8(y) == d.selectedY {
				screen.FillRect(posX, posY, 20, 20, 0xB6E0) //green
			}

			//highlight where it was before
			if d.moving && int8(x) == d.selectedPieceOriginalLocationX && int8(y) == d.selectedPieceOriginalLocationY {
				screen.FillRect(posX, posY, 20, 20, 0xFE60) // yellow highlight
			}

			piece := d.board[y][x]
			if piece == 0 {
				continue
			}

			// bit 3 is colour
			var color uint16
			if piece&0x08 == 0 {
				color = 0xFFFF // white
			} else {
				color = 0x0000 // black
			}

			// bits 0-2 are piece type
			var pieceLetter string
			switch piece & 0x07 {
			case 1:
				pieceLetter = "P"
			case 2:
				pieceLetter = "N"
			case 3:
				pieceLetter = "B"
			case 4:
				pieceLetter = "R"
			case 5:
				pieceLetter = "Q"
			case 6:
				pieceLetter = "K"
			}

			helpers.DrawText(screen, posX, posY, pieceLetter, color)
		}
	}
	screen.Present()
}
