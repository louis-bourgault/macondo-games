package chess

var pieceValues = [8]int{
	0,     // empty
	100,   // pawn
	320,   // knight
	330,   // bishop
	500,   // rook
	900,   // queen
	20000, // king
}

var pawnTable = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

var knightTable = [8][8]int{
	{-50, -40, -30, -30, -30, -30, -40, -50},
	{-40, -20, 0, 0, 0, 0, -20, -40},
	{-30, 0, 10, 15, 15, 10, 0, -30},
	{-30, 5, 15, 20, 20, 15, 5, -30},
	{-30, 0, 15, 20, 20, 15, 0, -30},
	{-30, 5, 10, 15, 15, 10, 5, -30},
	{-40, -20, 0, 5, 5, 0, -20, -40},
	{-50, -40, -30, -30, -30, -30, -40, -50},
}

var bishopTable = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}

var rookTable = [8][8]int{
	{0, 0, 0, 5, 5, 0, 0, 0},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}

var kingTable = [8][8]int{
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-20, -30, -30, -40, -40, -30, -30, -20},
	{-10, -20, -20, -20, -20, -20, -20, -10},
	{20, 20, 0, 0, 0, 0, 20, 20},
	{20, 30, 10, 0, 0, 10, 30, 20},
}

func chooseBestMove(d *ChessGame, depth int) (bestMove [4]int8, bestScore int) {
	return search(d, depth, -1000000, 1000000)
}

func search(d *ChessGame, depth int, alpha, beta int) ([4]int8, int) {
	if depth == 0 {
		score := d.evaluatePosition()
		if !d.playerTurn {
			score = -score
		}
		return [4]int8{}, score
	}

	moves := d.GenAllLegalMoves(d.playerTurn)
	if moves.count == 0 {
		if d.isKingInCheck(d.playerTurn) {
			return [4]int8{}, -99999 + depth
		}
		return [4]int8{}, 0
	}

	d.sortMoves(&moves)

	var bestMove [4]int8
	bestScore := -1000000

	for i := uint8(0); i < uint8(moves.count); i++ {
		move := moves.moves[i]

		captured := d.board[move[3]][move[2]]
		d.makeMove(move)
		d.playerTurn = !d.playerTurn

		_, score := search(d, depth-1, -beta, -alpha)
		score = -score

		d.playerTurn = !d.playerTurn
		d.unmakeMove(move)
		d.board[move[3]][move[2]] = captured

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}
	return bestMove, bestScore
}

func (d *ChessGame) isKingInCheck(whiteToMove bool) bool {
	var kx, ky int8
	for y := int8(0); y < 8; y++ {
		for x := int8(0); x < 8; x++ {
			p := d.board[y][x]
			if p&0x07 == 6 && (whiteToMove == (p&0x08 == 0)) {
				kx, ky = x, y
				break
			}
		}
	}
	return d.isSquareAttacked(kx, ky, !whiteToMove)
}

func (d *ChessGame) scoreMove(m [4]int8) int {
	movingPiece := d.board[m[1]][m[0]]
	capturedPiece := d.board[m[3]][m[2]]
	if capturedPiece != 0 {
		return 10*pieceValues[capturedPiece&0x07] - pieceValues[movingPiece&0x07]
	}
	return 0
}

func (d *ChessGame) sortMoves(moves *MoveList) {
	for i := 1; i < int(moves.count); i++ {
		j := i
		for j > 0 && d.scoreMove(moves.moves[j]) > d.scoreMove(moves.moves[j-1]) {
			moves.moves[j], moves.moves[j-1] = moves.moves[j-1], moves.moves[j]
			j--
		}
	}
}

func (d *ChessGame) evaluatePosition() int {
	score := 0
	for y := int8(0); y < 8; y++ {
		for x := int8(0); x < 8; x++ {
			piece := d.board[y][x]
			if piece == 0 {
				continue
			}
			pt := piece & 0x07
			value := pieceValues[pt]

			pstVal := 0
			switch pt {
			case 1:
				pstVal = pawnTable[y][x]
				if piece&0x08 != 0 {
					pstVal = pawnTable[7-y][x]
				}
			case 2:
				pstVal = knightTable[y][x]
				if piece&0x08 != 0 {
					pstVal = knightTable[7-y][x]
				}
			case 3:
				pstVal = bishopTable[y][x]
				if piece&0x08 != 0 {
					pstVal = bishopTable[7-y][x]
				}
			case 4:
				pstVal = rookTable[y][x]
				if piece&0x08 != 0 {
					pstVal = rookTable[7-y][x]
				}
			case 6:
				pstVal = kingTable[y][x]
				if piece&0x08 != 0 {
					pstVal = kingTable[7-y][x]
				}
			}

			if piece&0x08 == 0 {
				score += value + pstVal
			} else {
				score -= (value + pstVal)
			}
		}
	}
	return score
}
