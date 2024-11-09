package main

import (
	"fmt"
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

func (state *GameState) NewMove(postion Position, direction Direction, tiles Tiles, tilesScore *TilesScore, playerState *PlayerState) *Move {
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
	boardTile := state.tileBoard[pos.row][pos.column]
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
	if move == nil {

		// could not move ... pass

	}
	return move
}

func (state *GameState) AddPass(playerState *PlayerState) *Move {
	game := state.game
	options := game.options
	fmt := game.fmt
	state.consequtivePasses++
	move := state.NewMove(
		Position{game.height + 1, game.width + 1},
		EAST,
		Tiles{},
		&TilesScore{
			tileScores: TileScores{},
			multiplier: 0,
			score:      0,
		},
		&PlayerState{
			player:   playerState.player,
			playerNo: playerState.playerNo,
			score:    playerState.score,
			rack:     playerState.rack,
		})
	state.playerStates[move.playerState.playerNo] = move.playerState
	move.state = state
	state.move = move

	if options.debug > 0 {
		printState(state)
		fmt.Printf("AddPass :\n")
		printPlayer(state.game, playerState)
		fmt.Printf("\n")
	}
	return move
}

func (state *GameState) AddMove(partial *PartialMove, playerState *PlayerState) *Move {
	options := state.game.options
	corpus := state.game.corpus
	fmt := state.game.fmt
	state.consequtivePasses = 0
	move := state.NewMove(partial.startPos, partial.direction, partial.tiles, partial.score, &PlayerState{
		player:   playerState.player,
		playerNo: playerState.playerNo,
		score:    playerState.score + partial.score.score,
		rack:     partial.rack,
	})
	state.playerStates[move.playerState.playerNo] = move.playerState
	move.state = state
	state.move = move

	if options.debug > 0 {
		printState(state)
		fmt.Printf("AddMove :\n")
		printPartialMove(partial)
		printPlayer(state.game, playerState)
		fmt.Printf("\n")
	}

	pos := partial.startPos
	dir := partial.direction
	perpendicularOrientation := dir.Orientation().Perpendicular()
	var ok bool
	var nextPos Position
	for i, tile := range partial.tiles {
		boardTile := state.tileBoard[pos.row][pos.column]
		switch boardTile.kind {
		case TILE_EMPTY:
			state.tileBoard[pos.row][pos.column] = BoardTile{Tile: tile, validCrossLetters: NoValidCrossLetters}
			if options.debug > 0 {
				t := &state.tileBoard[pos.row][pos.column]
				fmt.Printf("   set tile %s = %s\n", pos.String(), t.String(corpus))
			}
			_, suffixPos := state.AdjacentPosition(pos, perpendicularOrientation.SuffixDirection())
			state.InvalidateValidCrossLetters(pos, suffixPos, perpendicularOrientation)

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
	state.InvalidateValidCrossLetters(partial.startPos, partial.endPos, dir.Orientation())

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
		t := &state.tileBoard[firstPrefixAnchor.row][firstPrefixAnchor.column]
		if options.debug > 0 {
			fmt.Printf("   invalidate %s prefix validCrossLetters %s : %s\n",
				perpendicularOrientation.String(), firstPrefixAnchor.String(),
				t.validCrossLetters[perpendicularOrientation].String(corpus))
		}
		t.validCrossLetters[perpendicularOrientation] = NullValidCrossLetters[perpendicularOrientation]
	}
	if state.game.IsValidPos(endPos) {
		if ok, firstSuffixAnchor := state.FindFirstAnchorFrom(endPos, orientation.SuffixDirection()); ok {
			t := &state.tileBoard[firstSuffixAnchor.row][firstSuffixAnchor.column]
			if options.debug > 0 {
				fmt.Printf("   invalidate %s suffix validCrossLetters %s : %s\n",
					perpendicularOrientation.String(), firstSuffixAnchor.String(),
					t.validCrossLetters[perpendicularOrientation].String(corpus))
			}
			t.validCrossLetters[perpendicularOrientation] = NullValidCrossLetters[perpendicularOrientation]
		}
	}
}
