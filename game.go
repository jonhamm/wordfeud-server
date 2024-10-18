package main

import (
	"math/rand"
	"slices"
)

const DW rune = '='
const TW rune = '#'
const DL rune = '+'
const TL rune = '*'
const EMPTY = '.'
const JOKER = '?'

const TL_COUNT = 12
const DL_COUNT = 24
const TW_COUNT = 8
const DW_COUNT = 16

const JOKER_COUNT = 2
const WIDTH = 15
const HEIGHT = 15

type SpecialField struct {
	kind  rune
	count int
}

type SpecialFields []SpecialField

var specialFields = SpecialFields{
	SpecialField{DW, DW_COUNT},
	SpecialField{TW, TW_COUNT},
	SpecialField{DL, DL_COUNT},
	SpecialField{TL, TL_COUNT},
}

type PieceValues map[rune]int8

type Move struct {
	player     *Player
	board      *Board
	row        int8
	column     int8
	horizontal bool
	word       Word
}
type MoveResult struct {
	fromBoard *Board
	move      *Move
	score     int
	toBoard   *Board
}

type Game struct {
	corpus *Corpus
	width  int8
	height int8
	values PieceValues
	pieces []rune
	moves  []MoveResult
}

func NewGame(corpus *Corpus, dimensions ...int8) *Game {
	var width int8
	var height int8
	switch len(dimensions) {
	case 0:
		width = WIDTH
		height = HEIGHT
	case 1:
		width = dimensions[0]
		height = width
	default:
		width = dimensions[0]
		height = dimensions[1]
	}
	game := Game{
		corpus: corpus,
		width:  width,
		height: height,
		values: PieceValues{},
		pieces: []rune{},
		moves:  make([]MoveResult, 0),
	}
	for _, piece := range corpus.pieces {
		game.values[piece.character] = piece.value
		game.pieces = slices.Grow(game.pieces, len(game.pieces)+int(piece.initalCount))
		var i int8
		for i = 0; i < piece.initalCount; i++ {
			game.pieces = append(game.pieces, piece.character)
		}
		for i = 0; i < JOKER_COUNT; i++ {
			game.pieces = append(game.pieces, JOKER)
		}
	}
	game.values[JOKER] = 0

	startBoard := NewBoard(&game)

	fillSpecialFields(startBoard)

	return &game
}

func fillSpecialFields(board *Board) {
	fillRandomSpecialFields(board)
}

func fillRandomSpecialFields(board *Board) {
	emptySquares := make([]Square, board.game.SquareCount())
	for _, f := range specialFields {
		for i := 0; i < f.count; i++ {
			n := rand.Intn(len(emptySquares))
			square := emptySquares[n]
			board.squares[square.row][square.column].content = f.kind
			emptySquares = slices.Delete(emptySquares, n, n+1)
		}
	}
}

func (game *Game) SquareCount() int {
	return int(game.width) * int(game.height)
}

func (game *Game) Width() int {
	return int(game.width)
}

func (game *Game) Height() int {
	return int(game.height)
}

func (game *Game) PieceCount() int {
	return len(game.pieces)
}

func (game *Game) DrawPiece() rune {
	i := rand.Intn(game.PieceCount())
	game.pieces = slices.Delete(game.pieces, i, i+1)
	piece := game.pieces[i]
	return piece
}

func MakeMoveResult(
	fromBoard *Board,
	move *Move,
	score int,
	toBoard *Board) *MoveResult {
	return &MoveResult{fromBoard, move, score, toBoard}
}

func MakeMove(player *Player,
	board *Board,
	row int8,
	column int8,
	horizontal bool,
	word Word) *Move {
	return &Move{player, board, row, column, horizontal, word}
}
