package main

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func (game *Game) play() bool {
	state := game.state
	playerNo := game.nextPlayer()
	playerState := state.playerStates[playerNo]
	//player := playerState.player
	newState := state.Move(playerState)
	if newState != nil {
		game.state = newState
		return true
	}
	return false
}

func (game *Game) nextPlayer() PlayerNo {
	if game.state.move == nil {
		return 0
	}
	for i, p := range game.players {
		if game.state.move.player == p {
			n := i + 1
			if n > len(game.players) {
				return 0
			}
			return PlayerNo(n)
		}
	}
	return 0
}

func (state *GameState) Move(playerState *PlayerState) *GameState {
	var p *message.Printer
	options := state.game.options
	if options.verbose {
		p = message.NewPrinter(language.Danish)
		p.Fprintf(options.out, "\n\nMove for player %v : %s\n", playerState.no, playerState.player.name)
		printState(options.out, state)
	}
	//anchors := state.GetAnchors()
	return nil
}

func (state *GameState) GetAnchors() Positions {
	anchors := make(Positions, 0)
	for r := Coordinate(0); r < state.game.height; r++ {
		for c := Coordinate(0); c < state.game.width; c++ {
			pos := Position{row: r, column: c}
			if state.tiles[r][c].kind == TILE_EMPTY {
				_, d, pos := state.AdjacentNonEmptyTile(pos)
				if d != DIRECTION_NONE {
					anchors = append(anchors, pos)
				}
			}
		}
	}
	if len(anchors) == 0 {
		anchors = append(anchors, Position{row: state.game.height / 2, column: state.game.width / 2})
	}
	return anchors
}

func (state *GameState) AdjacentNonEmptyTile(pos Position) (Tile, Direction, Position) {
	for _, d := range AllDirections {
		t, p := state.AdjacentTile(pos, d)
		switch t.kind {
		case TILE_JOKER, TILE_LETTER:
			return t, d, p
		}
	}
	return Tile{TILE_NONE, 0}, DIRECTION_NONE, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) AdjacentTile(pos Position, d Direction) (Tile, Position) {
	ok, adjacentPos := state.AdjacentPosition(pos, d)

	if ok {
		return state.tiles[adjacentPos.row][adjacentPos.column], adjacentPos
	}

	return Tile{TILE_NONE, 0}, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) AdjacentPosition(pos Position, d Direction) (bool, Position) {
	switch d {
	case DIRECTION_NORTH:
		if pos.row > 0 {
			return true, Position{pos.row - 1, pos.column}
		}
	case DIRECTION_SOUTH:
		if pos.row+1 < state.game.height {
			return true, Position{pos.row + 1, pos.column}

		}
	case DIRECTION_WEST:
		if pos.column > 0 {
			return true, Position{pos.row, pos.column}
		}

	case DIRECTION_EAST:
		if pos.column+1 < state.game.width {
			return true, Position{pos.row, pos.column + 1}

		}
	}
	return false, Position{state.game.height + 1, state.game.width + 1}
}