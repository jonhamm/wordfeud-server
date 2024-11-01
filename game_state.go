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
	DIRECTION_NONE  = Direction(0)
	DIRECTION_NORTH = Direction(1)
	DIRECTION_EAST  = Direction(2)
	DIRECTION_SOUTH = Direction(3)
	DIRECTION_WEST  = Direction(4)
)

var AllDirections = Directions{DIRECTION_NORTH, DIRECTION_SOUTH, DIRECTION_EAST, DIRECTION_WEST}

type Plane byte
type Planes []Plane

const (
	PLANE_HORIZONTAL = Plane(0)
	PLANE_VERTICAL   = Plane(1)
)

var AllPlanes = Planes{PLANE_HORIZONTAL, PLANE_VERTICAL}

const PlaneMax = PLANE_VERTICAL + 1

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

func (p Plane) Directions() Directions {
	switch p {
	case PLANE_HORIZONTAL:
		return Directions{DIRECTION_EAST, DIRECTION_WEST}
	case PLANE_VERTICAL:
		return Directions{DIRECTION_NORTH, DIRECTION_SOUTH}
	}
	panic("invalid plane (Plane.Directions)")
}

func (p Plane) Inverse() Plane {
	switch p {
	case PLANE_HORIZONTAL:
		return PLANE_VERTICAL
	case PLANE_VERTICAL:
		return PLANE_HORIZONTAL
	}
	panic("invalid plane (Plane.Inverse)")
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
	case DIRECTION_NORTH:
		return "N"
	case DIRECTION_SOUTH:
		return "S"
	case DIRECTION_EAST:
		return "E"
	case DIRECTION_WEST:
		return "W"
	}
	panic("invalid direction in Direction.String")
}
