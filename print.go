package main

import (
	"io"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func printBoard(f io.Writer, board *Board) {
	/*

		+--0--+--1--+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
		|*   4|     |     |     |     |     |     |     |     |     |     |     |     |     |     |
		0  K  1  .  |     |     |     |     |     |     |     |     |     |     |     |     |     |
		|   12|     |     |     |     |     |     |     |     |     |     |     |     |     |     |
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
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			p.Fprintf(f, "+--%d--", c%10)
		}
		p.Fprintf(f, "+\n")

		for c := 0; c < w; c++ {
			s := &squares[r][c]
			if s.content != 0 {
				score := board.game.values[s.content]
				p.Fprintf(f, "|%5d", score)
			} else {
				p.Fprintf(f, "     ")

			}
		}
		p.Fprintf(f, "|\n")

		for c := 0; c < w; c++ {
			s := &squares[r][c]
			p.Fprintf(f, "%d", r%10)
			if s.content != 0 {
				p.Fprintf(f, "  %c  ", s.content)
			}
		}
		p.Fprintf(f, "|\n")

		for c := 0; c < w; c++ {
			s := &squares[r][c]
			k := "  "
			switch s.kind {
			case DW:
				k = "dw"
			case TW:
				k = "tw"
			case DL:
				k = "dl"
			case TL:
				k = "tl"
			}
			p.Fprintf(f, "|   %s", k)

		}
		p.Fprintf(f, "|\n")

	}
}
