package main

import (
	"math/rand"
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
	players      []*Player
	state        *GameState
}

func NewGame(corpus *Corpus, players Players, dimensions ...Coordinate) *Game {
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
		players:      players,
		state:        nil,
	}
	game.board = NewBoard(&game)

	for _, tile := range corpus.tiles {
		game.letterScores[game.corpus.runeLetter[tile.character]] = tile.value
		game.tiles = slices.Grow(game.tiles, len(game.tiles)+int(tile.count))
		for i := byte(0); i < tile.count; i++ {
			game.tiles = append(game.tiles, Tile{TILE_LETTER, corpus.runeLetter[tile.character]})
		}
	}
	for i := 0; byte(i) < JOKER_COUNT; i++ {
		game.tiles = append(game.tiles, Tile{TILE_JOKER, 0})
	}

	game.state = InitialGameState(&game)

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
	switch tile.kind {
	case TILE_JOKER:
		return 0
	case TILE_NULL:
		return 0
	case TILE_LETTER:
		return game.letterScores[tile.letter]
	}
	return 0
}

func (game *Game) TakeTile() Tile {
	n := len(game.tiles)
	if n == 0 {
		return Tile{TILE_NULL, 0}
	}
	i := rand.Intn(n)
	t := game.tiles[i]
	game.tiles = slices.Delete(game.tiles, i, i+1)
	return t
}

func (game *Game) FillRack(rack Rack) Rack {

	for n := len(rack); n < int(RackSize); n++ {
		t := game.TakeTile()
		if t.kind == TILE_NULL {
			break
		}
		rack = append(rack, t)
	}
	return rack
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
