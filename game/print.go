package game

import (
	"fmt"
	"io"
	"os"
	"unicode"
	. "wordfeud/corpus"
)

func PrintBoard(board *Board, args ...string) {
	FprintBoard(os.Stdout, board, args...)
}

func FprintBoard(f io.Writer, board *Board, args ...string) {
	/*
		     0  1  2  3  4  5  6  7  8  9 10 11 12 13 14
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 0 |  |  |DL|  |  |DL|  |  |DW|  |DL|TL|  |DL|  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 1 |  |  |  |  |  |  |  |  |  |  |  |TL|  |  |DL|
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 2 |  |TW|DL|  |  |DL|  |  |DL|  |  |  |  |DW|  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 3 |TL|  |  |DW|  |  |  |  |  |  |TW|  |  |  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 4 |DL|  |TL|  |  |  |  |TL|DL|  |  |  |  |  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 5 |TW|  |  |DL|  |  |DL|  |  |  |DW|  |  |DL|  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 6 |  |DW|TL|  |  |  |  |  |DW|  |  |  |  |DW|DL|
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 7 |  |  |TW|  |  |  |  |  |DL|  |  |  |  |  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 8 |  |  |  |  |  |  |DW|TL|  |  |  |DW|TL|TL|  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		 9 |DW|DW|  |  |  |  |  |  |  |  |  |  |  |DL|  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		10 |  |  |DL|  |  |  |  |DW|  |DW|  |  |  |  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		11 |  |  |  |TW|  |DL|TL|DL|  |  |  |  |DL|  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		12 |  |TW|  |  |DL|TW|DW|  |  |  |DL|  |  |  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		13 |DW|  |  |  |  |TL|  |TW|DL|  |  |  |  |  |  |
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		14 |  |  |  |  |  |  |DW|  |DL|  |  |  |  |  |TL|
		   +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	p := board.game.board.game.fmt
	squares := board.squares
	w := board.game.Dimensions().Width
	h := board.game.Dimensions().Height
	p.Fprintf(f, "\n\n%s    ", indent)
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "%2d ", c)
	}
	p.Fprintf(f, "\n")
	for r := Coordinate(0); r < h; r++ {
		p.Fprintf(f, "%s   ", indent)
		for c := Coordinate(0); c < w; c++ {
			p.Fprintf(f, "+--")
		}
		p.Fprintf(f, "+\n")

		p.Fprintf(f, "%s%2d ", indent, r)

		for c := Coordinate(0); c < w; c++ {
			s := squares[r][c]
			k := "  "
			switch s {
			case DW:
				k = "DW"
			case TW:
				k = "TW"
			case DL:
				k = "DL"
			case TL:
				k = "TL"
			}
			p.Fprintf(f, "|%s", k)

		}
		p.Fprintf(f, "|\n")

	}
	p.Fprintf(f, "%s   ", indent)
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "+--")
	}
	p.Fprintf(f, "+\n")

}

func DebugState(state *GameState) {
	if state != nil {
		PrintState(state)
	}
}

func PrintState(state *GameState, args ...string) {
	FprintState(os.Stdout, state, args...)
}

func FprintState(f io.Writer, state *GameState, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	p := state.game.fmt
	board := state.game.board
	corpus := state.game.corpus
	squares := board.squares
	w := board.game.dimensions.Width
	h := board.game.dimensions.Height
	tiles := state.tileBoard
	placedMinRow := int(h)
	placedMaxRow := -1
	placedMinColumn := int(w)
	placedMaxColumn := -1

	if state.move != nil {

		for _, t := range state.move.tiles {
			p := t.pos
			if t.placedInMove {
				if int(p.row) > placedMaxRow {
					placedMaxRow = int(p.row)
				}
				if int(p.row) < placedMinRow {
					placedMinRow = int(p.row)
				}
			}
			if t.placedInMove {
				if int(p.column) > placedMaxColumn {
					placedMaxColumn = int(p.column)
				}
				if int(p.column) < placedMinColumn {
					placedMinColumn = int(p.column)
				}
			}
		}

	}

	p.Fprintf(f, "\n\n%s    ", indent)
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, " %2d   ", c)
	}
	p.Fprintf(f, "\n")
	for r := Coordinate(0); r < h; r++ {
		p.Fprintf(f, "%s   ", indent)
		for c := Coordinate(0); c < w; c++ {
			p.Fprintf(f, "+-----")
		}
		p.Fprintf(f, "+\n")

		p.Fprintf(f, "%s   ", indent)

		for c := Coordinate(0); c < w; c++ {
			letterScore := board.game.GetTileScore(tiles[r][c].Tile)
			s := squares[r][c]
			k := "  "
			switch s {
			case DW:
				k = "DW"
			case TW:
				k = "TW"
			case DL:
				k = "DL"
			case TL:
				k = "TL"
			case CE:
				k = "CE"
			}
			switch tiles[r][c].kind {
			case TILE_JOKER, TILE_LETTER:
				p.Fprintf(f, "|%s%3v", k, letterScore)
			default:
				p.Fprintf(f, "|%s   ", k)
			}
		}
		p.Fprintf(f, "|\n")

		p.Fprintf(f, "%s%2d ", indent, r)
		for c := Coordinate(0); c < w; c++ {
			t := tiles[r][c]
			l := ' '
			switch t.kind {
			case TILE_LETTER, TILE_JOKER:
				if t.letter != 0 {
					l = unicode.ToUpper(corpus.LetterToRune(t.letter))
				}

			}
			p.Fprintf(f, "|  %c  ", l)

		}
		p.Fprintf(f, "|\n")

		p.Fprintf(f, "%s   ", indent)

		for c := Coordinate(0); c < w; c++ {
			placedInMove := false
			if int(r) >= placedMinRow && int(r) <= placedMaxRow && int(c) >= placedMinColumn && int(c) <= placedMaxColumn {
				for _, t := range state.move.tiles {
					if t.pos.equal(Position{r, c}) {
						placedInMove = true
						break
					}
				}
			}
			if placedInMove {
				p.Fprintf(f, "|*   *")
			} else {
				p.Fprintf(f, "|     ")
			}
		}
		p.Fprintf(f, "|\n")
	}

	p.Fprintf(f, "%s   ", indent)
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "+-----")
	}
	p.Fprintf(f, "+\n")
	FprintPlayers(f, state.game, state.playerStates, indent)
	p.Fprintf(f, "current player : [%d]\n", state.playerNo)
	numberOfFreeTiles := len(state.freeTiles)
	numberOfRackTiles := state.NumberOfRackTiles()
	filledPositions := len(state.FilledPositions())
	p.Fprintf(f, "number of free tiles: %d\n", numberOfFreeTiles)
	p.Fprintf(f, "number of rack tiles: %d\n", numberOfRackTiles)
	p.Fprintf(f, "number of filled squares: %d\n", filledPositions)
	p.Fprintf(f, "total number of tiles: %d\n", filledPositions+numberOfFreeTiles+numberOfRackTiles)
}

func PrintStateOfGame(game Game, args ...string) {
	FprintStateOfGame(os.Stdout, game, args...)
}

func FprintStateOfGame(f io.Writer, game Game, args ...string) {
	_game := game._Game()
	if _game != nil && _game.state != nil {
		FprintState(f, _game.state, args...)
	}
}

func DebugPlayers(game Game, players PlayerStates) {
	if game != nil {
		PrintPlayers(game, players)
	}
}

func PrintPlayers(game Game, players PlayerStates, args ...string) {
	FprintPlayers(os.Stdout, game, players, args...)
}

func FprintPlayers(f io.Writer, game Game, players PlayerStates, args ...string) {
	indent := ""
	fmt := game.Fmt()
	if len(args) > 0 {
		indent = args[0]
	}
	if len(players) > 0 {
		fmt.Fprint(f, "\n\n")
	}
	for _, p := range players {
		fmt.Fprintf(f, "%s%s\n", indent, p.String(game.Corpus()))
	}

}

func PrintPlayer(game Game, player *PlayerState, args ...string) {
	FprintPlayer(os.Stdout, game, player, args...)
}

func DebugPlayer(game Game, player *PlayerState) {
	if game != nil && player != nil {
		PrintPlayer(game, player)
	}
}

func FprintPlayer(f io.Writer, game Game, player *PlayerState, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	p := game.Fmt()
	p.Fprintf(f, "%sPlayer %s\n", indent, player.String(game.Corpus()))
}

func DebugPartialMove(pm *PartialMove) {
	if pm != nil {
		PrintPartialMove(pm)
	}
}
func PrintPartialMove(pm *PartialMove, args ...string) {
	FprintPartialMove(os.Stdout, pm, args...)
}

func FprintPartialMove(f io.Writer, pm *PartialMove, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	state := pm.gameState
	game := state.game
	p := game.fmt
	corpus := game.corpus
	tiles := state.tileBoard
	word := pm.gameState.TilesToString(pm.tiles.Tiles())
	p.Fprintf(f, "%sPartialMove: %d  %s..%s \"%s\"\n", indent, pm.id, pm.startPos, pm.endPos, word)
	if game.IsValidPos(pm.startPos) {
		p.Fprintf(f, "%s   startPos:  %s   %s\n", indent, pm.startPos.String(), tiles[pm.startPos.row][pm.startPos.column].String(corpus))
	} else {
		p.Fprintf(f, "%s   startPos:  %s\n", indent, pm.startPos.String())
	}
	p.Fprintf(f, "%s   direction: %s\n", indent, pm.direction.String())
	if game.IsValidPos(pm.endPos) {
		p.Fprintf(f, "%s   endPos:  %s   %s\n", indent, pm.endPos.String(), tiles[pm.endPos.row][pm.endPos.column].String(corpus))
	} else {
		p.Fprintf(f, "%s   endPos:  %s\n", indent, pm.endPos.String())
	}
	p.Fprintf(f, "%s   rack:      %s\n", indent, pm.rack.String(corpus))
	p.Fprintf(f, "%s   tiles:     %s\n", indent, pm.tiles.String(corpus))
	p.Fprintf(f, "%s   word:      \"%s\"\n", indent, word)
	if pm.score != nil {
		p.Fprintf(f, "%s   score:     \n", indent)
		FprintMoveScore(f, pm.score, corpus, indent+"            ")
	}
	p.Fprintf(f, "%s   state:     \n", indent)
	pm.state.FprintState(f, indent+"            ")
}

func DebugPartialMoves(pms PartialMoves) {
	if pms != nil {
		PrintPartialMoves(pms)
	}

}

func PrintPartialMoves(pms PartialMoves, args ...string) {
	FprintPartialMoves(os.Stdout, pms, args...)

}

func FprintPartialMoves(f io.Writer, pms PartialMoves, args ...string) {
	if len(pms) == 0 {
		return
	}
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	p := pms[0].gameState.game.fmt
	p.Fprintf(f, "%sPartialMoves: \n", indent)
	indent += "    "
	for _, pm := range pms {
		FprintPartialMove(f, pm, indent)
	}
}

func DebugMove(pm *Move) {
	if pm != nil {
		PrintMove(pm)
	}
}

func PrintMove(pm *Move, args ...string) {
	FprintMove(os.Stdout, pm, args...)
}

func FprintMove(f io.Writer, move *Move, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	state := move.state
	game := state.game
	p := state.game.fmt
	corpus := state.game.corpus
	tiles := state.tileBoard
	word := move.state.TilesToString(move.tiles.Tiles())
	startPos := move.position
	endPos := startPos
	if game.IsValidPos(startPos) {
		_, endPos = state.AdjacentPosition(startPos, move.direction)
	}
	p.Fprintf(f, "%sMove: %d number %d  %s..%s \"%s\"\n",
		indent, move.id, move.seqno, startPos.String(), endPos.String(), word)
	if game.IsValidPos(move.position) {
		p.Fprintf(f, "%s   position:  %s   %s\n", indent, move.position.String(), tiles[move.position.row][move.position.column].String(corpus))
	} else {
		p.Fprintf(f, "%s   position:  %s\n", indent, move.position.String())
	}
	p.Fprintf(f, "%s   direction: %s\n", indent, move.direction.String())
	p.Fprintf(f, "%s   tiles:     %s\n", indent, move.tiles.String(corpus))
	p.Fprintf(f, "%s   word:      \"%s\"\n", indent, word)
	p.Fprintf(f, "%s   player:    %s\n", indent, move.playerState.String(corpus))
	if move.score != nil {
		p.Fprintf(f, "%s   score:    %s\n", indent, move.playerState.String(corpus))
		FprintMoveScore(f, move.score, corpus, indent+"              ")
	}
	p.Fprintf(f, "%s   state:     \n", indent)
	FprintState(f, move.state, indent+"             ")

}

func PrintMoveScore(ms *MoveScore, corpus Corpus, args ...string) {
	FprintMoveScore(os.Stdout, ms, corpus, args...)
}

func FprintMoveScore(f io.Writer, ms *MoveScore, corpus Corpus, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	fmt.Fprintf(f, "%sMoveScore: score: %d\n", indent, ms.score)
	FprintWordScores(f, ms.wordScores, corpus, indent+"   ")

}

func PrintWordScores(ws WordScores, corpus Corpus, args ...string) {
	FprintWordScores(os.Stdout, ws, corpus, args...)
}

func FprintWordScores(f io.Writer, ws WordScores, corpus Corpus, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}

	for _, s := range ws {
		FprintWordScore(f, s, corpus, indent)
	}
}

func PrintWordScore(ws *WordScore, corpus Corpus, args ...string) {
	FprintWordScore(os.Stdout, ws, corpus, args...)
}

func FprintWordScore(f io.Writer, ws *WordScore, corpus Corpus, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}

	pos := ""
	if len(ws.tileScores) > 0 {
		pos = ws.tileScores[0].tile.pos.String()
	}

	fmt.Fprintf(f, "%sTilesScore: \"%s\" %s %s score: %v word multiplier: %v\n", indent,
		ws.Word().String(corpus), ws.orientation.String(), pos, ws.score, ws.multiplier)
	for _, s := range ws.tileScores {
		FprintTileScore(f, &s, corpus, indent+"  ")
	}
}

func PrintTileScore(ts *TileScore, corpus Corpus, args ...string) {
	FprintTileScore(os.Stdout, ts, corpus, args...)
}

func FprintTileScore(f io.Writer, ts *TileScore, corpus Corpus, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
		fmt.Fprintf(f, "%sTileScore: %10s letterScore: %v multiplier: %v score: %4d \n",
			indent, ts.tile.String(corpus), ts.letterScore, ts.multiplier, ts.score)
	}
}
