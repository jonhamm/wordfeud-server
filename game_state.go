package main

import (
	"fmt"
	"strings"
)

type TileKind byte

const (
	TILE_EMPTY  = TileKind(0)
	TILE_JOKER  = TileKind(1)
	TILE_LETTER = TileKind(2)
	TILE_NONE   = TileKind(3)
)

type Direction byte
type Directions []Direction
type DirectionSet byte

const (
	NONE  = Direction(0)
	NORTH = Direction(1)
	SOUTH = Direction(2)
	WEST  = Direction(3)
	EAST  = Direction(4)
)

var AllDirections = Directions{NORTH, SOUTH, EAST, WEST}

type Orientation byte
type Planes []Orientation

const (
	HORIZONTAL = Orientation(0)
	VERTICAL   = Orientation(1)
)

var AllOrientations = Planes{HORIZONTAL, VERTICAL}

const PlaneMax = VERTICAL + 1

type Tile struct {
	kind   TileKind
	letter Letter
}

type Tiles []Tile

var NullTile = Tile{kind: TILE_NONE, letter: 0}

type ValidCrossLetters struct {
	ok      bool
	letters LetterSet
}

var NullValidCrossLetters = [PlaneMax]ValidCrossLetters{
	{ok: false, letters: 0},
	{ok: false, letters: 0},
}

type BoardTile struct {
	Tile
	validCrossLetters [PlaneMax]ValidCrossLetters
}

var NullBoardTile = BoardTile{
	Tile:              NullTile,
	validCrossLetters: NullValidCrossLetters,
}

type TileBoard [][]BoardTile
type GameState struct {
	game         *Game
	fromState    *GameState
	move         *Move
	tiles        TileBoard
	playerStates PlayerStates
}

const RackSize = 7

type Rack Tiles

type PlayerState struct {
	player *Player
	no     PlayerNo
	score  Score
	rack   Rack
}

type PlayerStates []PlayerState

func (o Orientation) Directions() Directions {
	switch o {
	case HORIZONTAL:
		return Directions{WEST, EAST}
	case VERTICAL:
		return Directions{NORTH, SOUTH}
	}
	panic("invalid plane (Orientation.Directions)")
}

func (o Orientation) PrefixDirection() Direction {
	return o.Directions()[0]
}

func (o Orientation) SuffixDirection() Direction {
	return o.Directions()[1]
}

func (o Orientation) Perpendicular() Orientation {
	switch o {
	case HORIZONTAL:
		return VERTICAL
	case VERTICAL:
		return HORIZONTAL
	}
	panic("invalid plane (Orientation.Perpendicular)")
}

func InitialGameState(game *Game) *GameState {
	state := &GameState{game: game, fromState: nil, move: nil, tiles: make(TileBoard, game.height)}
	allLetters := game.corpus.allLetters
	for r := Coordinate(0); r < game.height; r++ {
		state.tiles[r] = make([]BoardTile, game.width)
		for c := Coordinate(0); c < game.width; c++ {
			for p := range AllOrientations {
				validCrossLetters := &state.tiles[r][c].validCrossLetters[p]
				validCrossLetters.ok = true
				validCrossLetters.letters = allLetters
			}
		}
	}

	state.playerStates = make(PlayerStates, len(game.players))
	for i := 0; i < len(state.playerStates); i++ {
		state.playerStates[i] = PlayerState{
			player: game.players[i],
			no:     PlayerNo(i),
			score:  0,
			rack:   Rack{},
		}
		state.playerStates[i].rack = game.FillRack(state.playerStates[i].rack)
	}
	return state
}

func (directionSet DirectionSet) test(dir Direction) bool {
	return (directionSet & (1 << dir)) != 0
}

func (directionSet *DirectionSet) set(dir Direction) *DirectionSet {
	*directionSet |= DirectionSet(1 << dir)
	return directionSet
}

func (directionSet *DirectionSet) unset(dir Direction) *DirectionSet {
	*directionSet &^= DirectionSet(1 << dir)
	return directionSet
}

func (kind TileKind) String() string {
	switch kind {
	case TILE_EMPTY:
		return "="
	case TILE_JOKER:
		return "?"
	case TILE_LETTER:
		return "+"
	case TILE_NONE:
		return "-"
	}
	return "????"
}

func (dir Direction) String() string {
	switch dir {
	case NONE:
		return "NONE"
	case NORTH:
		return "N"
	case SOUTH:
		return "S"
	case EAST:
		return "E"
	case WEST:
		return "W"
	}
	return "????"
}

func (dirs Directions) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, dir := range dirs {
		if i > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(dir.String())
	}
	sb.WriteRune(']')
	return sb.String()
}

func (directionSet *DirectionSet) String(corpus *Corpus) string {
	var s strings.Builder
	var first = true
	s.WriteRune('{')
	for _, dir := range AllDirections {
		if directionSet.test(dir) {
			if first {
				first = false
			} else {
				s.WriteRune(',')
			}
			s.WriteString(dir.String())
		}
	}
	s.WriteRune('}')
	return s.String()
}

func (player *PlayerState) String(corpus *Corpus) string {
	return fmt.Sprintf("%v : %s score: %v rack: %v",
		player.no, player.player.name, player.score, player.rack.String(corpus))
}
