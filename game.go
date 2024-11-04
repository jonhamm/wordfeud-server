package main

import (
	"math/rand"
	"slices"

	"golang.org/x/text/message"
)

const JOKER_COUNT = 2
const WIDTH = 15
const HEIGHT = 15

type Score uint
type LetterScores [] /*Letter*/ Score
type tileBag []Tile

type Game struct {
	options      *GameOptions
	fmt          *message.Printer
	width        Coordinate
	height       Coordinate
	corpus       *Corpus
	dawg         *Dawg
	board        *Board
	letterScores LetterScores
	tiles        tileBag
	players      []*Player
	state        *GameState
}

func NewGame(options *GameOptions, players Players, dimensions ...Coordinate) (*Game, error) {
	var width Coordinate
	var height Coordinate
	printer := message.NewPrinter(options.language)

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
	var err error
	corpus, err := GetFileCorpus(GetLanguageFileName(options.language), GetLanguageAlphabet((options.language)))
	if err != nil {
		return nil, err
	}
	dawg, err := MakeDawg(corpus)
	if err != nil {
		return nil, err
	}
	game := Game{
		options:      options,
		width:        width,
		height:       height,
		corpus:       corpus,
		fmt:          printer,
		dawg:         dawg,
		board:        nil,
		letterScores: make(LetterScores, corpus.letterMax),
		tiles:        []Tile{},
		players:      players,
		state:        nil,
	}
	game.board = NewBoard(&game)

	for _, tile := range GetLanguageTiles(options.language) {
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

	return &game, nil
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
	case TILE_EMPTY:
		return 0
	case TILE_LETTER:
		return game.letterScores[tile.letter]
	}
	return 0
}

func (game *Game) CalcTileScore(position Position, tile Tile) Score {
	multiplier := Score(0)
	tileScore := game.GetTileScore(tile)
	switch game.board.squares[position.row][position.column] {
	case DL:
		multiplier = 2
	case TL:
		multiplier = 3
	}
	return multiplier * tileScore
}

func (game *Game) TakeTile() Tile {
	n := len(game.tiles)
	if n == 0 {
		return Tile{TILE_EMPTY, 0}
	}
	i := rand.Intn(n)
	t := game.tiles[i]
	game.tiles = slices.Delete(game.tiles, i, i+1)
	return t
}

func (game *Game) FillRack(rack Rack) Rack {

	for n := len(rack); n < int(RackSize); n++ {
		t := game.TakeTile()
		if t.kind == TILE_EMPTY {
			break
		}
		rack = append(rack, t)
	}
	return rack
}
