package main

import (
	"io"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func printBoard(f io.Writer, board *Board) {
	/*
		           0
				+-----+--1--+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
			0	|DW|  .  |     |     |     |     |     |     |     |     |     |     |     |     |     |
				+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				|     |     |     |     |     |     |     |     |     |     |     |     |     |     |     |
				+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+

	*/
	p := message.NewPrinter(language.Danish)
	squares := board.squares
	w := board.game.Width()
	h := board.game.Height()
	p.Fprintf(f, "    ")
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
