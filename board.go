package main

import (
	"fmt"
	"slices"
	"strings"
	. "wordfeud/corpus"

	"golang.org/x/text/language"
)

type Square byte
type Coordinate byte

type Position struct {
	row    Coordinate
	column Coordinate
}

type Positions []Position

const (
	DW Square = '='
	TW Square = '#'
	DL Square = '+'
	TL Square = '*'
	CE Square = '@'
)
const TL_COUNT = 12
const DL_COUNT = 24
const TW_COUNT = 8
const DW_COUNT = 16

type Board struct {
	game    *Game
	squares [][]Square
}

type SpecialField struct {
	kind  Square
	count int
}

type SpecialFields []SpecialField

var specialFields = SpecialFields{
	SpecialField{DW, DW_COUNT},
	SpecialField{TW, TW_COUNT},
	SpecialField{DL, DL_COUNT},
	SpecialField{TL, TL_COUNT},
}

func NewBoard(game *Game) *Board {
	board := Board{
		game:    game,
		squares: [][]Square{},
	}
	board.squares = make([][]Square, game.height)
	for i := range board.squares {
		board.squares[i] = make([]Square, game.width)
	}
	board.fillSpecialFields()
	return &board
}

func (board *Board) fillSpecialFields() {
	board.fillRandomSpecialFields()
}

func (board *Board) fillRandomSpecialFields() {
	normalSquares := make([]Position, board.game.SquareCount()-1)
	w := board.game.width
	h := board.game.height
	n := 0

	cr := h / 2
	cc := w / 2
	board.squares[cr][cc] = CE

	for r := Coordinate(0); r < h; r++ {
		for c := Coordinate(0); c < w; c++ {
			if board.squares[r][c] == 0 {
				normalSquares[n].row = r
				normalSquares[n].column = c
				n++
			}
		}
	}
	for _, f := range specialFields {
		for i := 0; i < f.count; i++ {
			n := board.game.rand.Intn(len(normalSquares))
			square := normalSquares[n]
			if board.squares[square.row][square.column] == 0 {
				board.squares[square.row][square.column] = f.kind
				normalSquares = slices.Delete(normalSquares, n, n+1)
			}
		}
	}
}

func (from Position) Distance(to Position, direction Direction) int {
	var length int
	switch direction {
	case EAST:
		length = int(to.column) - int(from.column)
	case WEST:
		length = int(from.column) - int(to.column)
	case SOUTH:
		length = int(to.row) - int(from.row)
	case NORTH:
		length = int(from.row) - int(to.row)
	}
	return length
}

func (pos Position) String() string {
	return fmt.Sprintf("(%v,%v)", pos.row, pos.column)
}

func (lhs Position) equal(rhs Position) bool {
	return lhs.row == rhs.row && lhs.column == rhs.column
}

func (pos Positions) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, p := range pos {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p.String())
	}
	sb.WriteRune(']')
	return sb.String()

}

func (rack Rack) String(corpus Corpus) string {
	var sb strings.Builder
	rack.Verify(corpus)
	sb.WriteString(fmt.Sprintf("(%d) [", len(rack)))
	for i, t := range rack {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(t.String(corpus))
	}
	sb.WriteRune(']')
	return sb.String()
}

func (rack Rack) Pretty(lang language.Tag, corpus Corpus) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(%d) [", len(rack)))
	for i, t := range rack {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(t.letter.String(corpus))
	}
	sb.WriteRune(']')
	return sb.String()
}

func (rack Rack) Verify(corpus Corpus) {
	for _, t := range rack {
		switch t.kind {
		case TILE_EMPTY, TILE_NONE:
			panic(fmt.Sprintf("invalid rack tile %s", t.String(corpus)))
		case TILE_JOKER, TILE_LETTER:
		}
	}
}
