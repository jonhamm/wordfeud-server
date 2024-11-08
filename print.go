package main

import (
	"fmt"
	"io"
	"os"
	"unicode"
)

func printOptions(options *GameOptions, args ...string) {
	fprintOptions(os.Stdout, options, args...)
}

func fprintOptions(f io.Writer, options *GameOptions, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	fmt.Fprintf(f, "%sGameOptions:\n", indent)
	fmt.Fprintf(f, "%s   cmd:         %s\n", indent, options.cmd)
	fmt.Fprintf(f, "%s   args:        %v\n", indent, options.args)
	fmt.Fprintf(f, "%s   help:        %v\n", indent, options.help)
	fmt.Fprintf(f, "%s   debug:       %v\n", indent, options.debug)
	fmt.Fprintf(f, "%s   ranSeed:     %v\n", indent, options.randSeed)
	fmt.Fprintf(f, "%s   count:       %v\n", indent, options.count)
	fmt.Fprintf(f, "%s   name:        %s\n", indent, options.name)
	fmt.Fprintf(f, "%s   language:    %s\n", indent, options.language.String())
	fmt.Fprintf(f, "%s   writeFile:   %v\n", indent, options.writeFile)
	fmt.Fprintf(f, "%s   directory:   %s\n", indent, options.directory)
	fmt.Fprintf(f, "%s   file:        %s\n", indent, options.file)
	fmt.Fprintf(f, "%s   fileFormat:  %s\n", indent, options.fileFormat.String())
}

func printBoard(board *Board, args ...string) {
	fprintBoard(os.Stdout, board, args...)
}

func fprintBoard(f io.Writer, board *Board, args ...string) {
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
	w := board.game.Width()
	h := board.game.Height()
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

func debugState(state *GameState) {
	printState(state)
}

func printState(state *GameState, args ...string) {
	fprintState(os.Stdout, state, args...)
}

func fprintState(f io.Writer, state *GameState, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	p := state.game.fmt
	board := state.game.board
	corpus := state.game.corpus
	squares := board.squares
	w := board.game.Width()
	h := board.game.Height()
	tiles := state.tiles
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
					l = unicode.ToUpper(corpus.letterRune[t.letter])
				}

			}
			p.Fprintf(f, "|  %c  ", l)

		}
		p.Fprintf(f, "|\n")

		p.Fprintf(f, "%s   ", indent)

		for c := Coordinate(0); c < w; c++ {
			p.Fprintf(f, "|     ")
		}
		p.Fprintf(f, "|\n")
	}

	p.Fprintf(f, "%s   ", indent)
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "+-----")
	}
	p.Fprintf(f, "+\n")
	fprintPlayers(f, state.game, state.playerStates, indent)

}

func debugPlayers(game *Game, players PlayerStates) {
	printPlayers(game, players)
}

func printPlayers(game *Game, players PlayerStates, args ...string) {
	fprintPlayers(os.Stdout, game, players, args...)
}

func fprintPlayers(f io.Writer, game *Game, players PlayerStates, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	if len(players) > 0 {
		game.fmt.Fprint(f, "\n\n")
	}
	for _, p := range players {
		fmt.Fprintf(f, "%s%s\n", indent, p.String(game.corpus))
	}

}

func printPlayer(game *Game, player *PlayerState, args ...string) {
	fprintPlayer(os.Stdout, game, player, args...)
}

func debugPlayer(game *Game, player *PlayerState) {
	printPlayer(game, player)
}

func fprintPlayer(f io.Writer, game *Game, player *PlayerState, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	p := game.fmt
	p.Fprintf(f, "%sPlayer %s\n", indent, player.String(game.corpus))
}

func debugPartialMove(pm *PartialMove) {
	printPartialMove(pm)
}
func printPartialMove(pm *PartialMove, args ...string) {
	fprintPartialMove(os.Stdout, pm, args...)
}

func fprintPartialMove(f io.Writer, pm *PartialMove, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	state := pm.gameState
	game := state.game
	p := game.fmt
	corpus := game.corpus
	tiles := state.tiles
	p.Fprintf(f, "%sPartialMove: %v\n", indent, pm.id)
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
	p.Fprintf(f, "%s   word:      \"%s\"\n", indent, pm.gameState.TilesToString(pm.tiles))
	if pm.score != nil {
		p.Fprintf(f, "%s   score:     \n", indent)
		fprintTilesScore(f, pm.score, corpus, indent+"            ")
	}
	p.Fprintf(f, "%s   state:     \n", indent)
	pm.gameState.game.dawg.fprintState(f, pm.state, indent+"            ")
}

func debugPartialMoves(pms PartialMoves) {
	printPartialMoves(pms)

}

func printPartialMoves(pms PartialMoves, args ...string) {
	fprintPartialMoves(os.Stdout, pms, args...)

}

func fprintPartialMoves(f io.Writer, pms PartialMoves, args ...string) {
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
		fprintPartialMove(f, pm, indent)
	}
}

func debugMove(pm *Move) {
	printMove(pm)
}

func printMove(pm *Move, args ...string) {
	fprintMove(os.Stdout, pm, args...)
}

func fprintMove(f io.Writer, move *Move, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	state := move.state
	game := state.game
	p := state.game.fmt
	corpus := state.game.corpus
	tiles := state.tiles
	p.Fprintf(f, "%sMove: %d number %d\n", indent, move.id, move.seqno)
	if game.IsValidPos(move.position) {
		p.Fprintf(f, "%s   position:  %s   %s\n", indent, move.position.String(), tiles[move.position.row][move.position.column].String(corpus))
	} else {
		p.Fprintf(f, "%s   position:  %s\n", indent, move.position.String())
	}
	p.Fprintf(f, "%s   direction: %s\n", indent, move.direction.String())
	p.Fprintf(f, "%s   tiles:     %s\n", indent, move.tiles.String(corpus))
	p.Fprintf(f, "%s   word:      \"%s\"\n", indent, move.state.TilesToString(move.tiles))
	p.Fprintf(f, "%s   player:    %s\n", indent, move.playerState.String(corpus))
	if move.score != nil {
		p.Fprintf(f, "%s   score:    %s\n", indent, move.playerState.String(corpus))
		fprintTilesScore(f, move.score, corpus, indent+"              ")
	}
	p.Fprintf(f, "%s   state:     \n", indent)
	fprintState(f, move.state, indent+"             ")

}

func printTilesScore(ts *TilesScore, corpus *Corpus, args ...string) {
	fprintTilesScore(os.Stdout, ts, corpus, args...)
}

func fprintTilesScore(f io.Writer, ts *TilesScore, corpus *Corpus, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}

	fmt.Fprintf(f, "%sTilesScore: score: %v word multiplier: %v\n", indent, ts.score, ts.multiplier)
	for _, s := range ts.tileScores {
		fprintTileScore(f, &s, corpus, indent+"  ")
	}
}

func printTileScore(ts *TileScore, corpus *Corpus, args ...string) {
	fprintTileScore(os.Stdout, ts, corpus, args...)
}

func fprintTileScore(f io.Writer, ts *TileScore, corpus *Corpus, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	placedInMove := ""
	if ts.placedInMove {
		placedInMove = "placed in move"
	}
	fmt.Fprintf(f, "%sTileScore: %10s letterScore: %v multiplier: %v score: %4d   %s\n",
		indent, ts.tile.String(corpus), ts.letterScore, ts.multiplier, ts.score, placedInMove)
}
