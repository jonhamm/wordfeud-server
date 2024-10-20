package main

import (
	"io"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func printBoard(f io.Writer, board *Board) {
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
	p := message.NewPrinter(language.Danish)
	squares := board.squares
	w := board.game.Width()
	h := board.game.Height()
	p.Fprintf(f, "\n\n    ")
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "%2d ", c)
	}
	p.Fprintf(f, "\n")
	for r := Coordinate(0); r < h; r++ {
		p.Fprintf(f, "   ")
		for c := Coordinate(0); c < w; c++ {
			p.Fprintf(f, "+--")
		}
		p.Fprintf(f, "+\n")

		p.Fprintf(f, "%2d ", r)

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
	p.Fprintf(f, "   ")
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "+--")
	}
	p.Fprintf(f, "+\n")

}

func printState(f io.Writer, state *GameState) {
	p := message.NewPrinter(language.Danish)
	board := state.game.board
	corpus := state.game.corpus
	squares := board.squares
	w := board.game.Width()
	h := board.game.Height()
	tiles := state.tiles
	p.Fprintf(f, "\n\n    ")
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, " %2d   ", c)
	}
	p.Fprintf(f, "\n")
	for r := Coordinate(0); r < h; r++ {
		p.Fprintf(f, "   ")
		for c := Coordinate(0); c < w; c++ {
			p.Fprintf(f, "+-----")
		}
		p.Fprintf(f, "+\n")

		p.Fprintf(f, "   ")

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
			p.Fprintf(f, "|%s   ", k)
		}
		p.Fprintf(f, "|\n")

		p.Fprintf(f, "%2d ", r)
		for c := Coordinate(0); c < w; c++ {
			t := tiles[r][c]
			l := ' '
			switch t.kind {
			case TILE_LETTER:
			case TILE_JOKER:
				if t.letter != 0 {
					l = corpus.letterRune[l]
				}

			}
			p.Fprintf(f, "|  %c  ", l)

		}
		p.Fprintf(f, "|\n")

		p.Fprintf(f, "   ")

		for c := Coordinate(0); c < w; c++ {
			p.Fprintf(f, "|     ")
		}
		p.Fprintf(f, "|\n")
	}

	p.Fprintf(f, "   ")
	for c := Coordinate(0); c < w; c++ {
		p.Fprintf(f, "+-----")
	}
	p.Fprintf(f, "+\n")
	printPlayers(f, state.game, state.playerStates)

}

func printPlayers(f io.Writer, game *Game, players PlayerStates) {
	p := message.NewPrinter(language.Danish)
	if len(players) > 0 {
		p.Fprint(f, "\n\n")
	}
	for _, p := range players {
		printPlayer(f, game, p)
	}

}

func printPlayer(f io.Writer, game *Game, player *PlayerState) {
	p := message.NewPrinter(language.Danish)
	p.Fprintf(f, "Player %v : %s\n", player.no, player.player.name)
	p.Fprintf(f, "   Score: %v\n", player.score)
	p.Fprintf(f, "    Rack: ")
	for _, r := range player.rack {
		switch r.kind {
		case TILE_LETTER:
			p.Fprintf(f, "%c(%v) ", game.corpus.letterRune[r.letter], game.letterScores[r.letter])
		case TILE_JOKER:
			p.Fprintf(f, "?(0) ")
		case TILE_NULL:
			p.Fprintf(f, "NULL")
		}
	}
	p.Fprint(f, "\n")

}
