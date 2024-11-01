package main

import "strings"

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
	EAST  = Direction(2)
	SOUTH = Direction(3)
	WEST  = Direction(4)
)

var AllDirections = Directions{NORTH, SOUTH, EAST, WEST}

type Orientation byte
type Planes []Orientation

const (
	HORIZONTAL = Orientation(0)
	VERTICAL   = Orientation(1)
)

var AllPlanes = Planes{HORIZONTAL, VERTICAL}

const PlaneMax = VERTICAL + 1

type Tile struct {
	kind   TileKind
	letter Letter
}

var NullTile = Tile{kind: TILE_NONE, letter: 0}

type ValidCrossLetters struct {
	ok      bool
	letters LetterSet
}
type BoardTile struct {
	Tile
	validCrossLetters [PlaneMax]ValidCrossLetters
	score             Score
}

var NullBoardTile = BoardTile{
	Tile: NullTile,
	validCrossLetters: [PlaneMax]ValidCrossLetters{
		{ok: false, letters: 0},
		{ok: false, letters: 0},
	},
	score: 0,
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

type Rack []Tile

type PlayerState struct {
	player *Player
	no     PlayerNo
	score  Score
	rack   Rack
}

type PlayerStates []*PlayerState

func (o Orientation) Directions() Directions {
	switch o {
	case HORIZONTAL:
		return Directions{EAST, WEST}
	case VERTICAL:
		return Directions{NORTH, SOUTH}
	}
	panic("invalid plane (Orientation.Directions)")
}

func (o Orientation) Inverse() Orientation {
	switch o {
	case HORIZONTAL:
		return VERTICAL
	case VERTICAL:
		return HORIZONTAL
	}
	panic("invalid plane (Orientation.Inverse)")
}

func InitialGameState(game *Game) *GameState {
	state := &GameState{game: game, fromState: nil, move: nil, tiles: make(TileBoard, game.height)}
	allLetters := game.corpus.allLetters
	for r := Coordinate(0); r < game.height; r++ {
		state.tiles[r] = make([]BoardTile, game.width)
		for c := Coordinate(0); c < game.width; c++ {
			for p := range AllPlanes {
				validCrossLetters := &state.tiles[r][c].validCrossLetters[p]
				validCrossLetters.ok = true
				validCrossLetters.letters = allLetters
			}
		}
	}

	state.playerStates = make(PlayerStates, len(game.players))
	for i := 0; i < len(state.playerStates); i++ {
		state.playerStates[i] = &PlayerState{
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

func (dir Direction) String() string {
	switch dir {
	case NORTH:
		return "N"
	case SOUTH:
		return "S"
	case EAST:
		return "E"
	case WEST:
		return "W"
	}
	panic("invalid direction in Direction.String")
}
