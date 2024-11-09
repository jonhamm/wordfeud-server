package main

import (
	"math/rand"

	"golang.org/x/text/message"
)

const JOKER_COUNT = 2
const WIDTH = 15
const HEIGHT = 15

type Score uint
type LetterScores [] /*Letter*/ Score

type Game struct {
	options       *GameOptions
	seqno         int
	randSeed      uint64
	rand          *rand.Rand
	fmt           *message.Printer
	width         Coordinate
	height        Coordinate
	corpus        *Corpus
	dawg          *Dawg
	board         *Board
	letterScores  LetterScores
	players       []*Player
	state         *GameState
	nextMoveSeqNo uint
	nextMoveId    uint
}

func NewGame(options *GameOptions, seqno int, players Players, dimensions ...Coordinate) (*Game, error) {
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
	dawg, err := NewDawg(corpus)
	if err != nil {
		return nil, err
	}
	game := Game{
		options:       options,
		seqno:         seqno,
		width:         width,
		height:        height,
		corpus:        corpus,
		fmt:           printer,
		dawg:          dawg,
		board:         nil,
		letterScores:  make(LetterScores, corpus.letterMax),
		players:       make(Players, len(players)+1),
		state:         nil,
		nextMoveSeqNo: 1,
		nextMoveId:    1,
	}

	if options.debug > 0 && options.count <= 1 {
		game.randSeed = options.randSeed
		game.rand = options.rand
	} else {
		game.randSeed = options.rand.Uint64()
		game.rand = rand.New(rand.NewSource(int64(game.randSeed)))
	}

	game.players[0] = SystemPlayer
	copy(game.players[1:], players)

	game.board = NewBoard(&game)

	if options.debug > 0 {
		printer.Printf("****** New Game %s-%d ******  randSeed: %v\n", game.options.name, seqno, game.randSeed)
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

func (game *Game) TilesToWord(tiles Tiles) Word {
	word := make(Word, 0, len(tiles))
	for _, t := range tiles {
		switch t.kind {
		case TILE_LETTER, TILE_JOKER:
			word = append(word, t.letter)
		}
	}
	return word
}

func (game *Game) WordToTiles(word Word) Tiles {
	tiles := make(Tiles, 0, len(word))
	for _, letter := range word {
		tiles = append(tiles, Tile{kind: TILE_LETTER, letter: letter})
	}
	return tiles
}

func (game *Game) WordToRack(word Word) Rack {
	rack := make(Rack, 0, len(word))
	for _, letter := range word {
		rack = append(rack, Tile{kind: TILE_LETTER, letter: letter})
	}
	return rack
}

func (game *Game) NextMoveId() uint {
	id := game.nextMoveId
	game.nextMoveId++
	return id
}

func (game *Game) NextMoveSeqNo() uint {
	seqno := game.nextMoveSeqNo
	game.nextMoveSeqNo++
	return seqno
}

func (game *Game) CollectStates() GameStates {
	return game.state.CollectStates()
}

func (game *Game) IsValidPos(pos Position) bool {
	return pos.row < game.height && pos.column < game.width
}
