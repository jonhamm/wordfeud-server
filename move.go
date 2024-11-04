package main

import "fmt"

type Move struct {
	state       *GameState
	playerState PlayerState
	position    Position
	direction   Direction
	tiles       []Tile
	score       Score
}

func (state *GameState) MakeMove(postion Position, direction Direction, tiles []Tile, playerState PlayerState) *Move {
	move := &Move{state, playerState, postion, direction, tiles, 0}
	move.score = state.CalcScore(postion, direction, tiles)
	return move
}

func (state *GameState) CalcScore(anchor Position, dir Direction, tiles []Tile) Score {
	score := Score(0)
	multiplyer := Score(1)
	squares := state.game.board.squares
	pos := anchor
	boardTile := state.tiles[pos.row][pos.column]
	for _, t := range tiles {
		var nextPos Position
		letterScore := state.game.GetTileScore(t)
		if boardTile.kind == TILE_EMPTY {
			// this tile was placed in this move
			// use the square modifiers DL, TL, DW, TW
			switch squares[pos.row][pos.column] {
			case DL:
				letterScore *= 2
			case TL:
				letterScore *= 3
			case DW:
				multiplyer *= 2
			case TW:
				multiplyer *= 3
			}
		}
		score += letterScore
		boardTile, nextPos = state.AdjacentTile(pos, dir)
		if boardTile.kind == TILE_NONE {
			panic(fmt.Sprintf("invalid next position after %s in direction %s (GameState.CalcScore)", pos.String(), dir.String()))
		}
		pos = nextPos
	}

	score *= multiplyer

	return score
}
