package game

import (
	"math/rand"
	. "wordfeud/context"
	. "wordfeud/corpus"
	. "wordfeud/dawg"

	"golang.org/x/text/message"
)

const JOKER_COUNT = 2
const WIDTH = 15
const HEIGHT = 15

const MaxConsequtivePasses = 3

type Score uint
type LetterScores [] /*Letter*/ Score
type Dimensions struct {
	Width  Coordinate
	Height Coordinate
}

type Game interface {
	Options() *GameOptions
	Corpus() Corpus
	Fmt() *message.Printer
	Dimensions() Dimensions
	LetterScores() LetterScores
	Board() *Board
	SquareCount() int
	Play() bool
	rand() *rand.Rand
	_Game() *_Game
}

type _Game struct {
	options       *GameOptions
	seqno         int
	RandSeed      uint64
	_rand         *rand.Rand
	fmt           *message.Printer
	dimensions    Dimensions
	corpus        Corpus
	dawg          Dawg
	board         *Board
	letterScores  LetterScores
	players       []*Player
	state         *GameState
	nextMoveSeqNo uint
	nextMoveId    uint
}

func NewGame(options *GameOptions, seqno int, players Players, dimensions ...Coordinate) (Game, error) {
	var width Coordinate
	var height Coordinate
	printer := message.NewPrinter(options.Language)

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
	corpus, err := NewCorpus(options.Language)
	if err != nil {
		return nil, err
	}
	content, err := corpus.GetLanguageContent()
	if err != nil {
		return nil, err
	}
	dawg, err := NewDawg(content, options.Options)
	if err != nil {
		return nil, err
	}

	game := &_Game{
		options:       options,
		seqno:         seqno,
		dimensions:    Dimensions{Width: width, Height: height},
		corpus:        corpus,
		fmt:           printer,
		dawg:          dawg,
		board:         nil,
		letterScores:  make(LetterScores, corpus.LetterMax()),
		players:       make(Players, len(players)+1),
		state:         nil,
		nextMoveSeqNo: 1,
		nextMoveId:    1,
	}

	if options.Debug > 0 && options.Count <= 1 {
		game.RandSeed = options.RandSeed
		game._rand = options.Rand
	} else {
		game.RandSeed = options.Rand.Uint64()
		game._rand = rand.New(rand.NewSource(int64(game.RandSeed)))
	}

	game.players[0] = SystemPlayer
	copy(game.players[1:], players)

	game.board = NewBoard(game)

	if options.Debug > 0 {
		printer.Printf("****** New *_Game %s-%d ******  RandSeed: %v\n", game.options.Name, seqno, game.RandSeed)
	}

	game.state, err = initialGameState(game)

	return game, err
}

func (game *_Game) _Game() *_Game {
	return game
}

func (game *_Game) Options() *GameOptions {
	return game.options
}

func (game *_Game) Board() *Board {
	return game.board
}
func (game *_Game) Corpus() Corpus {
	return game.corpus
}

func (game *_Game) rand() *rand.Rand {
	return game._rand
}

func (game *_Game) SquareCount() int {
	return int(game.dimensions.Width) * int(game.dimensions.Height)
}

func (game *_Game) Dimensions() Dimensions {
	return game.dimensions
}

func (game *_Game) LetterScores() LetterScores {
	return game.letterScores
}

func (game *_Game) Fmt() *message.Printer {
	return game.fmt
}

func (game *_Game) GetTileScore(tile Tile) Score {
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

func (game *_Game) CalcTileScore(position Position, tile Tile) Score {
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

func (game *_Game) TilesToWord(tiles Tiles) Word {
	word := make(Word, 0, len(tiles))
	for _, t := range tiles {
		switch t.kind {
		case TILE_LETTER, TILE_JOKER:
			word = append(word, t.letter)
		}
	}
	return word
}

func (game *_Game) WordToTiles(word Word) Tiles {
	tiles := make(Tiles, 0, len(word))
	for _, letter := range word {
		tiles = append(tiles, Tile{kind: TILE_LETTER, letter: letter})
	}
	return tiles
}

func (game *_Game) WordToRack(word Word) Rack {
	rack := make(Rack, 0, len(word))
	for _, letter := range word {
		rack = append(rack, Tile{kind: TILE_LETTER, letter: letter})
	}
	return rack
}

func (game *_Game) NextMoveId() uint {
	id := game.nextMoveId
	game.nextMoveId++
	return id
}

func (game *_Game) NextMoveSeqNo() uint {
	seqno := game.nextMoveSeqNo
	game.nextMoveSeqNo++
	return seqno
}

func (game *_Game) CollectStates() GameStates {
	return game.state.CollectStates()
}

func (game *_Game) IsValidPos(pos Position) bool {
	return pos.row < game.dimensions.Height && pos.column < game.dimensions.Width
}
