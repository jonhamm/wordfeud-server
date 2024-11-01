package main

import (
	"math/rand"
	"slices"
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
	normalSquares := make([]Position, board.game.SquareCount())
	w := board.game.width
	h := board.game.height
	n := 0

	for r := Coordinate(0); r < h; r++ {
		for c := Coordinate(0); c < w; c++ {
			normalSquares[n].row = r
			normalSquares[n].column = c
			n++
		}
	}
	for _, f := range specialFields {
		for i := 0; i < f.count; i++ {
			n := rand.Intn(len(normalSquares))
			square := normalSquares[n]
			if square.row != h/2 && square.column != w/2 {
				board.squares[square.row][square.column] = f.kind
				normalSquares = slices.Delete(normalSquares, n, n+1)
			}
		}
	}
}

func (board *Board) CalcTileScore(position Position, tile Tile) Score {
	multiplier := Score(0)
	tileScore := board.game.GetTileScore(tile)
	switch board.squares[position.row][position.column] {
	case DL:
		multiplier = 2
	case TL:
		multiplier = 3
	}
	return multiplier * tileScore
}
