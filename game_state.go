package main

type TileKind byte

const (
	TILE_EMPTY  = TileKind(0)
	TILE_JOKER  = TileKind(1)
	TILE_LETTER = TileKind(2)
	TILE_NONE   = TileKind(3)
)

type Direction byte

const (
	DIRECTION_NONE  = Direction(0)
	DIRECTION_NORTH = Direction(1)
	DIRECTION_EAST  = Direction(2)
	DIRECTION_SOUTH = Direction(3)
	DIRECTION_WEST  = Direction(4)
)

var AllDirections = []Direction{DIRECTION_NORTH, DIRECTION_EAST, DIRECTION_SOUTH, DIRECTION_WEST}

type Tile struct {
	kind   TileKind
	letter Letter
}

var NullTile = Tile{kind: TILE_NONE, letter: 0}

type BoardTile struct {
	Tile
	validHorizontal   LetterSet
	validVertical     LetterSet
	validHorizontalOk bool
	validVerticalOk   bool
	score             Score
}

var NullBoardTile = BoardTile{
	Tile:              NullTile,
	validHorizontal:   0,
	validVertical:     0,
	validHorizontalOk: false,
	validVerticalOk:   false,
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

func InitialGameState(game *Game) *GameState {
	state := &GameState{game: game, fromState: nil, move: nil, tiles: make(TileBoard, game.height)}
	for r := Coordinate(0); r < game.height; r++ {
		state.tiles[r] = make([]BoardTile, game.width)
		for c := Coordinate(0); r < game.width; c++ {
			state.tiles[r][c].validHorizontal = game.corpus.allLetters
			state.tiles[r][c].validVertical = game.corpus.allLetters
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
