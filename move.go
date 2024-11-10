package main

import (
	"fmt"
	"slices"
	"strings"
)

type TileScore struct {
	tile        MoveTile
	letterScore Score
	multiplier  Score
	score       Score
}

type TileScores []TileScore
type WordScore struct {
	tileScores  TileScores
	orientation Orientation
	multiplier  Score
	score       Score
}

type WordScores []*WordScore

type MoveScore struct {
	wordScores WordScores
	score      Score
}

type MoveTile struct {
	Tile
	pos          Position
	placedInMove bool
}

type MoveTiles []MoveTile
type Move struct {
	id          uint
	seqno       uint
	state       *GameState
	playerState *PlayerState
	position    Position
	direction   Direction
	tiles       MoveTiles
	score       *MoveScore
}

func (state *GameState) NewMove(position Position, direction Direction, tiles MoveTiles, moveScore *MoveScore, playerState *PlayerState) *Move {
	move := &Move{
		id:          state.NextMoveId(),
		seqno:       state.game.NextMoveSeqNo(),
		state:       state,
		playerState: playerState,
		position:    position,
		direction:   direction,
		tiles:       tiles,
		score:       moveScore,
	}

	if move.score == nil {
		move.score = state.CalcScore(tiles, direction.Orientation())
	}

	return move
}

func (state *GameState) CalcScore(tiles MoveTiles, orientation Orientation) *MoveScore {
	moveScore := MoveScore{
		wordScores: make(WordScores, 0, 1),
		score:      0,
	}

	score := state.CalcWordScore(tiles, orientation)
	if score == nil {
		panic("CalcWordScore returned nil for main score")
	}

	moveScore.score += score.score
	moveScore.wordScores = append(moveScore.wordScores, score)

	perpendicular := orientation.Perpendicular()

	for _, tile := range tiles {
		if tile.placedInMove {
			wordTiles := state.GetWordMoveTiles(tile.pos, tile, perpendicular)
			if len(wordTiles) > 1 {
				score = state.CalcWordScore(wordTiles, perpendicular)
				if score == nil {
					panic("CalcWordScore returned nil")
				}
				moveScore.score += score.score
				moveScore.wordScores = append(moveScore.wordScores, score)
			}
		}
	}
	return &moveScore
}

func (state *GameState) CalcWordScore(tiles MoveTiles, orientation Orientation) *WordScore {
	wordScore := WordScore{
		tileScores:  make(TileScores, len(tiles)),
		orientation: orientation,
		multiplier:  1,
	}
	squares := state.game.board.squares
	for i, tile := range tiles {
		pos := tile.pos
		tileScore := &wordScore.tileScores[i]
		tileScore.tile = tile
		tileScore.multiplier = 1
		tileScore.letterScore = state.game.GetTileScore(tile.Tile)
		if tile.placedInMove {
			// this tile was placed in this move
			// use the square modifiers DL, TL, DW, TW
			switch squares[pos.row][pos.column] {
			case DL:
				tileScore.multiplier *= 2
			case TL:
				tileScore.multiplier *= 3
			case DW:
				wordScore.multiplier *= 2
			case TW:
				wordScore.multiplier *= 3
			}
		}
		tileScore.score += tileScore.multiplier * tileScore.letterScore
		wordScore.score += tileScore.score
	}

	wordScore.score *= wordScore.multiplier

	return &wordScore
}

func (state *GameState) GetWordMoveTiles(pos Position, tile MoveTile, orientation Orientation) MoveTiles {
	wordTiles := make(MoveTiles, 0)
	tiles := state.tileBoard
	prefixPos := pos
	dir := orientation.PrefixDirection()
	for {
		ok, p := state.AdjacentPosition(prefixPos, dir)
		if !ok {
			break
		}
		tile := &tiles[p.row][p.column]
		if tile.kind == TILE_EMPTY {
			break
		}
		if tile.kind == TILE_NONE {
			panic(fmt.Sprintf("unexpected TILE_NONE on GameState board[%v,%v] (GameState.FindPrefix)", p.row, p.column))
		}
		wordTiles = append(wordTiles, MoveTile{Tile: tile.Tile, pos: p, placedInMove: false})
		prefixPos = p
	}

	slices.Reverse(wordTiles)

	dir = dir.Reverse()
	wordTiles = append(wordTiles, tile)

	suffixPos := pos
	for {
		ok, p := state.AdjacentPosition(suffixPos, dir)
		if !ok {
			break
		}
		tile := &tiles[p.row][p.column]
		if tile.kind == TILE_EMPTY {
			break
		}
		if tile.kind == TILE_NONE {
			panic(fmt.Sprintf("unexpected TILE_NONE on GameState board[%v,%v] (GameState.FindPrefix)", p.row, p.column))
		}
		wordTiles = append(wordTiles, MoveTile{Tile: tile.Tile, pos: p, placedInMove: false})
		suffixPos = p
	}
	return wordTiles
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
		MoveTiles{},
		&MoveScore{
			wordScores: WordScores{},
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
		fmt.Printf("AddMove %d : %s..%s \"%s\"\n", move.seqno, partial.startPos, partial.endPos, state.TilesToString(partial.tiles.Tiles()))
		printPartialMove(partial)
		printPlayer(state.game, playerState)
		fmt.Printf("\n")
	}

	pos := partial.startPos
	dir := partial.direction
	perpendicularOrientation := dir.Orientation().Perpendicular()
	for _, tile := range partial.tiles {
		pos = tile.pos
		boardTile := state.tileBoard[pos.row][pos.column]
		switch boardTile.kind {
		case TILE_EMPTY, TILE_NONE:
			if !tile.placedInMove {
				panic(fmt.Sprintf("move generation does not place tile %s at %s which is  empty %s (GameState.AddMove)",
					tile.String(corpus), pos.String(), boardTile.Tile.String(corpus)))
			}
			state.tileBoard[pos.row][pos.column] = BoardTile{Tile: tile.Tile, validCrossLetters: NoValidCrossLetters}
			if options.debug > 0 {
				t := &state.tileBoard[pos.row][pos.column]
				fmt.Printf("   set tile %s = %s\n", pos.String(), t.String(corpus))
			}
			_, suffixPos := state.AdjacentPosition(pos, perpendicularOrientation.SuffixDirection())
			state.InvalidateValidCrossLetters(pos, suffixPos, perpendicularOrientation)

		case TILE_JOKER, TILE_LETTER:
			if tile.placedInMove {
				panic(fmt.Sprintf("move generation has tile %s at %s which is not empty (GameState.AddMove)",
					tile.String(corpus), pos.String()))
			} else {
				if !tile.Tile.equal(boardTile.Tile) {
					panic(fmt.Sprintf("move generation has tile %s at %s which is not empty and differs %s (GameState.AddMove)",
						tile.String(corpus), pos.String(), boardTile.Tile.String(corpus)))
				}
			}
		// this is a tile from a previous move -- skip
		default:
			panic(fmt.Sprintf("move generation will add new tile of unknown kind %d (GameState.AddMove)", tile.kind))
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

func (tile MoveTile) String(corpus *Corpus) string {
	placedInMove := '-'
	if tile.placedInMove {
		placedInMove = '+'
	}
	return fmt.Sprintf("%s %c '%s':%v:%s",
		tile.pos.String(), placedInMove, tile.letter.String(corpus), tile.letter, tile.kind.String())

}

func (tiles MoveTiles) String(corpus *Corpus) string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, tile := range tiles {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(tile.String(corpus))
	}
	sb.WriteRune(']')
	return sb.String()
}

func (moveTiles MoveTiles) Tiles() Tiles {
	tiles := make(Tiles, len(moveTiles))
	for i, t := range moveTiles {
		tiles[i] = t.Tile
	}
	return tiles
}
