package main

import (
	"math/rand"
	"slices"
)

const DW byte = '='
const TW byte = '#'
const DL byte = '+'
const TL byte = '*'
const EMPTY rune = '.'
const JOKER rune = '?'

const TL_COUNT = 12
const DL_COUNT = 24
const TW_COUNT = 8
const DW_COUNT = 16

const JOKER_COUNT = 2
const WIDTH = 15
const HEIGHT = 15

type SpecialField struct {
	kind  byte
	count int
}

type SpecialFields []SpecialField

var specialFields = SpecialFields{
	SpecialField{DW, DW_COUNT},
	SpecialField{TW, TW_COUNT},
	SpecialField{DL, DL_COUNT},
	SpecialField{TL, TL_COUNT},
}

type PieceValues map[rune]byte

type Move struct {
	player     *Player
	board      *Board
	row        byte
	column     byte
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
	width  byte
	height byte
	values PieceValues
	pieces []rune
	moves  []*MoveResult
}

func NewGame(corpus *Corpus, dimensions ...byte) *Game {
	var width byte
	var height byte
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
		moves:  make([]*MoveResult, 0),
	}
	for _, piece := range corpus.pieces {
		game.values[piece.character] = piece.value
		game.pieces = slices.Grow(game.pieces, len(game.pieces)+int(piece.initalCount))
		var i byte
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
	game.moves = append(game.moves, MakeMoveResult(nil, nil, 0, startBoard))

	return &game
}

func fillSpecialFields(board *Board) {
	fillRandomSpecialFields(board)
}

func fillRandomSpecialFields(board *Board) {
	normalSquares := make([]*Square, board.game.SquareCount())
	w := board.game.Width()
	h := board.game.Height()
	n := 0
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			normalSquares[n] = &board.squares[r][c]
			n++
		}
	}
	for _, f := range specialFields {
		for i := 0; i < f.count; i++ {
			n := rand.Intn(len(normalSquares))
			square := normalSquares[n]
			board.squares[square.row][square.column].kind = f.kind
			normalSquares = slices.Delete(normalSquares, n, n+1)
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
	row byte,
	column byte,
	horizontal bool,
	word Word) *Move {
	return &Move{player, board, row, column, horizontal, word}
}
