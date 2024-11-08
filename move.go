package main

import (
	"fmt"
	"slices"
)

type TileScore struct {
	tile         Tile
	placedInMove bool
	letterScore  Score
	multiplier   Score
	score        Score
}

type TileScores []TileScore
type TilesScore struct {
	tileScores TileScores
	multiplier Score
	score      Score
}
type Move struct {
	id          uint
	seqno       uint
	state       *GameState
	playerState *PlayerState
	position    Position
	direction   Direction
	tiles       Tiles
	score       *TilesScore
}

func (state *GameState) MakeMove(postion Position, direction Direction, tiles Tiles, tilesScore *TilesScore, playerState *PlayerState) *Move {
	move := &Move{state.NextMoveId(), state.game.NextMoveSeqNo(), state, playerState, postion, direction, tiles, nil}
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
			tileScore.placedInMove = true
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

func (state *GameState) Move(playerState *PlayerState) *Move {
	fmt := state.game.fmt
	options := state.game.options
	if options.verbose {
		fmt.Fprintf(options.out, "\n\nMove for player %v : %s\n", playerState.player.id, playerState.player.name)
		fprintState(options.out, state)
	}
	state.PrepareMove()

	filteredPartialMoves := state.FilterBestMove(state.GenerateAllMoves(playerState))

	if len(filteredPartialMoves) == 0 {
		return nil
	}

	move := state.AddMove(filteredPartialMoves[0], playerState)
	return move
}

func (state *GameState) AddMove(partial *PartialMove, playerState *PlayerState) *Move {
	options := state.game.options
	corpus := state.game.corpus
	fmt := state.game.fmt
	playerNo := playerState.player.id
	move := state.MakeMove(partial.startPos, partial.direction, partial.tiles, partial.score, &PlayerState{
		player: playerState.player,
		score:  playerState.score + partial.score.score,
		rack:   partial.rack,
	})
	newState := &GameState{
		game:         state.game,
		fromState:    state,
		move:         move,
		playerStates: slices.Concat(state.playerStates[:playerNo], PlayerStates{move.playerState}, state.playerStates[playerNo+1:]),
	}
	height := state.game.height
	width := state.game.width
	newState.tiles = make([][]BoardTile, height)

	if options.debug > 0 {
		printState(state)
		fmt.Printf("AddMove :\n")
		printPartialMove(partial)
		printPlayer(newState.game, playerState)
		fmt.Printf("\n")
	}

	newState.playerStates[move.playerState.player.id] = move.playerState

	for r := Coordinate(0); r < height; r++ {
		newState.tiles[r] = make([]BoardTile, width)
	}
	for r := Coordinate(0); r < height; r++ {
		for c := Coordinate(0); c < width; c++ {
			newState.tiles[r][c] = state.tiles[r][c]
		}
	}
	pos := partial.startPos
	dir := partial.direction
	perpendicularOrientation := dir.Orientation().Perpendicular()
	var ok bool
	var nextPos Position
	for i, tile := range partial.tiles {
		boardTile := newState.tiles[pos.row][pos.column]
		switch boardTile.kind {
		case TILE_EMPTY:
			newState.tiles[pos.row][pos.column] = BoardTile{Tile: tile, validCrossLetters: NoValidCrossLetters}
			if options.debug > 0 {
				t := &newState.tiles[pos.row][pos.column]
				fmt.Printf("   set tile %s = %s\n", pos.String(), t.String(corpus))
			}
			_, suffixPos := state.AdjacentPosition(pos, perpendicularOrientation.SuffixDirection())
			newState.InvalidateValidCrossLetters(pos, suffixPos, perpendicularOrientation)

		case TILE_JOKER, TILE_LETTER:
			if !boardTile.Tile.equal(tile) {
				panic(fmt.Sprintf("move generation will add new tile %s at %s which is not empty and differs %s (GameState.AddMove)",
					tile.String(corpus), pos.String(), boardTile.Tile.String(corpus)))
			}
			// this is a tile from a previous move -- skip
		case TILE_NONE:
			panic(fmt.Sprintf("move generation will add new tile at non-existing next position after %s direction %s (GameState.AddMove)", nextPos.String(), dir.String()))
		default:
			panic(fmt.Sprintf("move generation will add new tile of unknown kind %d (GameState.AddMove)", tile.kind))
		}

		if i+1 < len(partial.tiles) {
			ok, nextPos = state.AdjacentPosition(pos, dir)
			if !ok {
				panic(fmt.Sprintf("move generation will add new tile at non-existing next position after %s direction %s (GameState.AddMove)", nextPos.String(), dir.String()))
			}
			pos = nextPos
		}
	}
	newState.InvalidateValidCrossLetters(partial.startPos, partial.endPos, dir.Orientation())
	move.state = newState
	if options.debug > 0 {
		fmt.Printf("AddMove complete :\n")
		printMove(move)
		fmt.Printf("\n")
	}
	return move
}

func (state *GameState) InvalidateValidCrossLetters(startPos Position, endPos Position, orientation Orientation) {
	options := state.game.options
	corpus := state.game.corpus
	perpendicularOrientation := orientation.Perpendicular()
	if options.debug > 0 {
		fmt.Printf("InvalidateValidCrossLetters %s %s %s\n", startPos.String(), endPos.String(), orientation.String())
	}
	if ok, firstPrefixAnchor := state.FindFirstAnchorAfter(startPos, orientation.PrefixDirection()); ok {
		t := &state.tiles[firstPrefixAnchor.row][firstPrefixAnchor.column]
		if options.debug > 0 {
			fmt.Printf("   invalidate %s prefix validCrossLetters %s : %s\n",
				perpendicularOrientation.String(), firstPrefixAnchor.String(),
				t.validCrossLetters[perpendicularOrientation].String(corpus))
		}
		t.validCrossLetters[perpendicularOrientation] = NullValidCrossLetters[perpendicularOrientation]
	}
	if state.game.IsValidPos(endPos) {
		if ok, firstSuffixAnchor := state.FindFirstAnchorFrom(endPos, orientation.SuffixDirection()); ok {
			t := &state.tiles[firstSuffixAnchor.row][firstSuffixAnchor.column]
			if options.debug > 0 {
				fmt.Printf("   invalidate %s suffix validCrossLetters %s : %s\n",
					perpendicularOrientation.String(), firstSuffixAnchor.String(),
					t.validCrossLetters[perpendicularOrientation].String(corpus))
			}
			t.validCrossLetters[perpendicularOrientation] = NullValidCrossLetters[perpendicularOrientation]
		}
	}

}
