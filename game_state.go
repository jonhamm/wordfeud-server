package main

type TileKind byte

const (
	TILE_NULL   = TileKind(0)
	TILE_JOKER  = TileKind(1)
	TILE_LETTER = TileKind(2)
)

type Tile struct {
	kind   TileKind
	letter Letter
}

type TileBoard [][]Tile
type GameState struct {
	game         *Game
	fromState    *GameState
	move         *Move
	tiles        TileBoard
	playerStates PlayerStates
}

type Move struct {
	game       *Game
	player     *Player
	position   Position
	horizontal bool
	word       Word
	score      Score
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
		state.tiles[r] = make([]Tile, game.width)
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
