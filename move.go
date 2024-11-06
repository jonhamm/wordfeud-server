package main

import "fmt"

type TileScore struct {
	tile        Tile
	letterScore Score
	multiplier  Score
	score       Score
}

type TileScores []TileScore
type TilesScore struct {
	tileScores TileScores
	multiplier Score
	score      Score
}
type Move struct {
	id          uint
	state       *GameState
	playerState PlayerState
	position    Position
	direction   Direction
	tiles       Tiles
	score       *TilesScore
}

func (state *GameState) MakeMove(postion Position, direction Direction, tiles Tiles, tilesScore *TilesScore, playerState PlayerState) *Move {
	move := &Move{state.NextMoveId(), state, playerState, postion, direction, tiles, nil}
	move.score = tilesScore
	if move.score == nil {
		move.score = state.CalcScore(postion, direction, tiles)
	}
	return move
}

func (state *GameState) CalcScore(anchor Position, dir Direction, tiles Tiles) *TilesScore {
	tilesScore := TilesScore{
		tileScores: make(TileScores, len(tiles)),
		multiplier: 1,
		score:      0,
	}
	squares := state.game.board.squares
	pos := anchor
	boardTile := state.tiles[pos.row][pos.column]
	for i, t := range tiles {
		var nextPos Position
		tileScore := &tilesScore.tileScores[i]
		tileScore.tile = t
		tileScore.multiplier = 1
		tileScore.letterScore = state.game.GetTileScore(t)
		if boardTile.kind == TILE_EMPTY {
			// this tile was placed in this move
			// use the square modifiers DL, TL, DW, TW
			switch squares[pos.row][pos.column] {
			case DL:
				tileScore.multiplier *= 2
			case TL:
				tileScore.multiplier *= 3
			case DW:
				tilesScore.multiplier *= 2
			case TW:
				tilesScore.multiplier *= 2
			}
		}
		tileScore.score += tileScore.multiplier * tileScore.letterScore
		tilesScore.score += tileScore.score

		if i+1 < len(tiles) {
			boardTile, nextPos = state.AdjacentTile(pos, dir)
			if boardTile.kind == TILE_NONE {
				panic(fmt.Sprintf("invalid next position after %s in direction %s (GameState.CalcScore)", pos.String(), dir.String()))
			}
			pos = nextPos
		}
	}

	tilesScore.score *= tilesScore.multiplier

	return &tilesScore
}
