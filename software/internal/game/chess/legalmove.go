package chess

type MoveList struct {
	moves [256][4]int8
	count int
}

func (m *MoveList) add(fromX, fromY, toX, toY int8) {
	m.moves[m.count] = [4]int8{fromX, fromY, toX, toY}
	m.count++
}

func isOwn(piece uint8, whiteToMove bool) bool {
	if piece == 0 {
		return false
	}
	if whiteToMove {
		return piece&0x08 == 0
	}
	return piece&0x08 != 0
}

var diagDirs = [4][2]int8{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
var straightDirs = [4][2]int8{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

func (d *ChessGame) GenAllLegalMoves(whiteToMove bool) MoveList {
	var pseudo MoveList
	for y := int8(0); y < 8; y++ {
		for x := int8(0); x < 8; x++ {
			piece := d.board[y][x]
			if piece == 0 {
				continue
			}
			if whiteToMove != (piece&0x08 == 0) {
				continue
			}
			d.genPseudoMoves(&pseudo, x, y, whiteToMove)
		}
	}
	var legal MoveList
	for i := uint8(0); i < uint8(pseudo.count); i++ {
		if d.isMoveLegal(pseudo.moves[i], whiteToMove) {
			m := pseudo.moves[i]
			legal.add(m[0], m[1], m[2], m[3])
		}
	}
	return legal
}

func (d *ChessGame) genPseudoMoves(moves *MoveList, x, y int8, whiteToMove bool) {
	switch d.board[y][x] & 0x07 {
	case 1:
		dir := int8(-1)
		startRow := int8(6)
		if !whiteToMove {
			dir = 1
			startRow = 1
		}
		ny := y + dir
		if ny >= 0 && ny < 8 && d.board[ny][x] == 0 {
			moves.add(x, y, x, ny)
			ny2 := y + 2*dir
			if y == startRow && d.board[ny2][x] == 0 {
				moves.add(x, y, x, ny2)
			}
		}
		for _, dx := range [2]int8{-1, 1} {
			nx := x + dx
			if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
				target := d.board[ny][nx]
				if target != 0 && !isOwn(target, whiteToMove) {
					moves.add(x, y, nx, ny)
				}
			}
		}
	case 2:
		for _, o := range [8][2]int8{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
			nx, ny := x+o[0], y+o[1]
			if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 && !isOwn(d.board[ny][nx], whiteToMove) {
				moves.add(x, y, nx, ny)
			}
		}
	case 3:
		d.addSlidingMoves(moves, x, y, whiteToMove, &diagDirs)
	case 4:
		d.addSlidingMoves(moves, x, y, whiteToMove, &straightDirs)
	case 5:
		d.addSlidingMoves(moves, x, y, whiteToMove, &diagDirs)
		d.addSlidingMoves(moves, x, y, whiteToMove, &straightDirs)
	case 6:
		for dy := int8(-1); dy <= 1; dy++ {
			for dx := int8(-1); dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 && !isOwn(d.board[ny][nx], whiteToMove) {
					moves.add(x, y, nx, ny)
				}
			}
		}
	}
}

func (d *ChessGame) addSlidingMoves(moves *MoveList, x, y int8, whiteToMove bool, dirs *[4][2]int8) {
	for _, dir := range dirs {
		nx, ny := x+dir[0], y+dir[1]
		for nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			target := d.board[ny][nx]
			if isOwn(target, whiteToMove) {
				break
			}
			moves.add(x, y, nx, ny)
			if target != 0 {
				break
			}
			nx += dir[0]
			ny += dir[1]
		}
	}
}

func (d *ChessGame) isMoveLegal(m [4]int8, whiteToMove bool) bool {
	piece := d.board[m[1]][m[0]]
	captured := d.board[m[3]][m[2]]
	d.board[m[3]][m[2]] = piece
	d.board[m[1]][m[0]] = 0

	var kx, ky int8
	for y := int8(0); y < 8; y++ {
		for x := int8(0); x < 8; x++ {
			p := d.board[y][x]
			if p&0x07 == 6 && (whiteToMove == (p&0x08 == 0)) {
				kx, ky = x, y
			}
		}
	}

	inCheck := d.isSquareAttacked(kx, ky, whiteToMove)

	d.board[m[1]][m[0]] = piece
	d.board[m[3]][m[2]] = captured
	return !inCheck
}

func (d *ChessGame) isSquareAttacked(x, y int8, byBlack bool) bool {
	var enemyColor uint8
	if byBlack {
		enemyColor = 0x08
	}

	for _, o := range [8][2]int8{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
		nx, ny := x+o[0], y+o[1]
		if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			p := d.board[ny][nx]
			if p&0x07 == 2 && p&0x08 == enemyColor {
				return true
			}
		}
	}

	for _, dir := range diagDirs {
		nx, ny := x+dir[0], y+dir[1]
		for nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			p := d.board[ny][nx]
			if p != 0 {
				pt := p & 0x07
				if p&0x08 == enemyColor && (pt == 3 || pt == 5) {
					return true
				}
				break
			}
			nx += dir[0]
			ny += dir[1]
		}
	}

	for _, dir := range straightDirs {
		nx, ny := x+dir[0], y+dir[1]
		for nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			p := d.board[ny][nx]
			if p != 0 {
				pt := p & 0x07
				if p&0x08 == enemyColor && (pt == 4 || pt == 5) {
					return true
				}
				break
			}
			nx += dir[0]
			ny += dir[1]
		}
	}

	pawnDir := int8(1)
	if byBlack {
		pawnDir = -1
	}
	py := y + pawnDir
	if py >= 0 && py < 8 {
		for _, dx := range [2]int8{-1, 1} {
			px := x + dx
			if px >= 0 && px < 8 {
				p := d.board[py][px]
				if p&0x07 == 1 && p&0x08 == enemyColor {
					return true
				}
			}
		}
	}

	for dy := int8(-1); dy <= 1; dy++ {
		for dx := int8(-1); dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
				p := d.board[ny][nx]
				if p&0x07 == 6 && p&0x08 == enemyColor {
					return true
				}
			}
		}
	}

	return false
}

// func genLegalMovesForPiece(board [][]uint8, x, y int8) [][4]int8 {
// 	//the return value is an array of moves, where each move is an array of 4 int8s: [fromX, fromY, toX, toY]
// 	var moves [][4]int8
// 	switch board[x][y] & 0x07 {
// 	case 1:
// 		//pawn
// 		if board[x][y+1] != 0 {
// 			moves = append(moves, [4]int8{x, y, x, y + 1})
// 		}
// 	case 2:
// 		//knight
// 	case 3:
// 		//bishop
// 	case 4:
// 		//rook
// 	case 5:
// 		//queen
// 	case 6:
// 		//king
// 	}
// 	for i := 0; i < len(moves); i++ {
// 		if !verifyMove(board, moves[i]) {
// 			moves = append(moves[:i], moves[i+1:]...)
// 			i--
// 		}
// 	}
// 	return moves

// }

// func verifyMove(board [][]uint8, move [4]int8) bool {
// 	//does some more complicated things, like checking for checks

// 	return true

// }
