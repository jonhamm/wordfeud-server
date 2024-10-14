package main

const DW = '='
const TW = '#'
const DL = '+'
const TL = '*'
const EMPTY = '.'

type Square struct {
	row     int
	column  int
	content rune
}

type PieceValues map[rune]int

type Board struct {
	corpus     *Corpus
	width      int
	height     int
	values     PieceValues
	pieces     []rune
	pieceCount map[rune]int
	squares    [][]Square
}

func NewBoard(corpus *Corpus, width int, height int) *Board {
	board := Board{
		corpus:     corpus,
		width:      width,
		height:     height,
		values:     PieceValues{},
		pieces:     []rune{},
		pieceCount: map[rune]int{},
		squares:    [][]Square{},
	}
	board.squares = make([][]Square, height)
	for i := range board.squares {
		board.squares[i] = make([]Square, width)
	}
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			square := &board.squares[r][c]
			square.row = r
			square.column = c
			square.content = EMPTY
		}
	}
	return &board
}
