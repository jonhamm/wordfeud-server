package main

import (
	"slices"
)

const JOKER_COUNT = 2
const WIDTH = 15
const HEIGHT = 15

type Score uint
type LetterScores [] /*Letter*/ Score
type tileBag []Tile

type Game struct {
	corpus       *Corpus
	width        Coordinate
	height       Coordinate
	board        *Board
	letterScores LetterScores
	tiles        tileBag
}

func NewGame(corpus *Corpus, dimensions ...Coordinate) *Game {
	var width Coordinate
	var height Coordinate
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
		corpus:       corpus,
		width:        width,
		height:       height,
		board:        nil,
		letterScores: make(LetterScores, corpus.letterMax),
		tiles:        []Tile{},
	}
	game.board = NewBoard(&game)

	for _, tile := range corpus.tiles {
		game.letterScores[game.corpus.runeLetter[tile.character]] = tile.value
		game.tiles = slices.Grow(game.tiles, len(game.tiles)+int(tile.count))
		var i byte
		for i = 0; i < tile.count; i++ {
			game.tiles = append(game.tiles, Tile{false, corpus.runeLetter[tile.character]})
		}
		for i = 0; i < JOKER_COUNT; i++ {
			game.tiles = append(game.tiles, Tile{true, 0})
		}
	}

	//game.moves = append(game.moves, MakeMoveResult(nil, nil, 0, startBoard))

	return &game
}

func (game *Game) SquareCount() int {
	return int(game.width) * int(game.height)
}

func (game *Game) Width() Coordinate {
	return game.width
}

func (game *Game) Height() Coordinate {
	return game.height
}

func (game *Game) GetTileScore(tile Tile) Score {
	if tile.joker {
		return 0
	}
	return game.letterScores[tile.letter]
}

/*
func (game *Game) PieceCount() int {
	return len(game.tiles)
}

func (game *Game) DrawPiece() rune {
	i := rand.Intn(game.PieceCount())
	game.tiles = slices.Delete(game.tiles, i, i+1)
	piece := game.tiles[i]
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
*/
