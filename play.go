package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type RackTile struct {
	tile Tile
	rack Rack
}

type RackTiles []RackTile

type PartialMove struct {
	id        uint
	gameState *GameState
	rack      Rack
	startPos  Position
	endPos    Position
	direction Direction
	state     DawgState
	tiles     MoveTiles
	score     *MoveScore
}

type PartialMoves []*PartialMove

func (game *Game) Play() bool {
	options := game.options
	curState := game.state
	playerNo := curState.NextPlayer()
	curPlayerStates := curState.playerStates
	curPlayerState := curPlayerStates[playerNo]
	playerState := &PlayerState{
		player:   curPlayerState.player,
		playerNo: playerNo,
		score:    curPlayerState.score,
		rack:     slices.Clone(curPlayerState.rack),
	}
	state := &GameState{
		game:              game,
		fromState:         curState,
		tileBoard:         curState.tileBoard.Clone(),
		move:              nil,
		playerStates:      slices.Concat(curPlayerStates[:playerNo], PlayerStates{playerState}, curPlayerStates[playerNo+1:]),
		playerNo:          playerNo,
		freeTiles:         slices.Clone(curState.freeTiles),
		consequtivePasses: curState.consequtivePasses,
	}
	move := state.Move(playerState)
	if move != nil {
		state.move = move
		game.state = state // == move.state

		for _, ps := range game.state.playerStates {
			state.FillRack(ps)
		}
		if options.debug > 0 {
			game.fmt.Printf("game play completed move : %s\n", playerState.String(game.corpus))
		}
		for _, ps := range game.state.playerStates {
			if ps.playerNo != NoPlayer && ps.NumberOfRackTiles() == 0 {
				if options.verbose {
					fmt.Printf("Player %d %s has no more tiles in rack - game completed\n", ps.player.id, ps.player.name)
				}
				return false
			}
		}
		if state.consequtivePasses >= MaxConsequtivePasses {
			if options.verbose {
				fmt.Printf("There has been %d conequtive passes - game completed\n", state.consequtivePasses)
			}
			return false
		}

		if options.writeFile {
			gameFileName, err := WriteGameFile(game)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing game file \"%s\"\n%v\n", gameFileName, err.Error())
				return false
			}
			if options.verbose {
				fmt.Printf("wrote game file after move %d \"%s\"\n", game.nextMoveSeqNo-1, gameFileName)
			}
		}
		return true
	}
	return false
}

func (state *GameState) NextPlayer() PlayerNo {
	if state.move == nil {
		return 1
	}
	n := state.playerNo + 1
	if int(n) >= len(state.playerStates) {
		return 1
	}
	return n
}

func (state *GameState) GetAnchors(coordinate Coordinate, orientation Orientation) Positions {
	options := state.game.options
	anchors := make(Positions, 0)
	switch orientation {
	case HORIZONTAL:
		for c := Coordinate(0); c < state.game.width; c++ {
			switch state.tileBoard[coordinate][c].kind {
			case TILE_EMPTY:
				pos := Position{row: coordinate, column: c}
				if state.IsAnchor(pos) {
					anchors = append(anchors, pos)
				}
			}
		}

	case VERTICAL:
		for r := Coordinate(0); r < state.game.height; r++ {
			switch state.tileBoard[r][coordinate].kind {
			case TILE_EMPTY:
				pos := Position{row: r, column: coordinate}
				if state.IsAnchor(pos) {
					anchors = append(anchors, pos)
				}
			}
		}
	}
	if options.debug > 0 {
		fmt := state.game.fmt
		fmt.Printf("GetAnchors %d %s : %s\n", coordinate, orientation.String(), anchors.String())
	}
	return anchors
}

func (state *GameState) FindFirstAnchorAfter(position Position, direction Direction) (bool, Position) {
	var pos Position
	var ok bool
	for ok, pos = state.AdjacentPosition(position, direction); ok; ok, pos = state.AdjacentPosition(pos, direction) {
		if state.IsAnchor(pos) {
			return true, pos
		}
	}
	return false, pos
}

func (state *GameState) FindFirstAnchorFrom(position Position, direction Direction) (bool, Position) {
	var pos Position
	var ok bool
	for ok, pos = state.game.IsValidPos(position), position; ok; ok, pos = state.AdjacentPosition(pos, direction) {
		if state.IsAnchor(pos) {
			return true, pos
		}
	}
	return false, pos
}

func (state *GameState) PrepareMove() {
	game := state.game
	tiles := state.tileBoard
	h := game.height
	w := game.width
	for r := Coordinate(0); r < h; r++ {
		for c := Coordinate(0); c < w; c++ {
			for _, orientation := range AllOrientations {
				validCrossLetters := &tiles[r][c].validCrossLetters[orientation]
				if !validCrossLetters.ok {
					validLetters := state.CalcValidCrossLetters(Position{r, c}, orientation)
					if game.options.debug > 0 {
						fmt.Printf("validate %s %s validCrossLetters %s\n",
							Position{r, c}.String(), orientation.String(), validLetters.String(game.corpus))
					}

					validCrossLetters.letters = validLetters
					validCrossLetters.ok = true

				}
			}
		}
	}

}

func (state *GameState) GenerateAllMoves(playerState *PlayerState) PartialMoves {
	options := state.game.options
	corpus := state.game.corpus
	if options.debug > 0 {
		fmt.Printf("\n\n--------------------------------\n GenrateAllMoves: player: %s\n", playerState.String(corpus))
		printState(state)
	}
	playerState.rack.Verify(corpus)
	out := make(PartialMoves, 0, 100)
	fmt := state.game.fmt
	width := state.game.width
	height := state.game.height
	for r := Coordinate(0); r < height; r++ {
		anchors := state.GetAnchors(r, HORIZONTAL)
		if options.debug > 0 {
			if len(anchors) > 0 {
				fmt.Fprintf(options.out, "Anchors row %v: %s\n", r, anchors.String())
			}
		}
		for _, anchor := range anchors {
			moves := state.GenerateAllMovesForAnchor(playerState, anchor, HORIZONTAL)
			out = slices.Concat(out, moves)
		}
	}
	for c := Coordinate(0); c < width; c++ {
		anchors := state.GetAnchors(c, VERTICAL)
		if options.verbose {
			if len(anchors) > 0 {
				fmt.Fprintf(options.out, "Anchors column %d %s\n", c, anchors.String())
			}
		}
		for _, anchor := range anchors {
			moves := state.GenerateAllMovesForAnchor(playerState, anchor, VERTICAL)
			out = slices.Concat(out, moves)
		}
	}
	return out
}

func (state *GameState) GenerateAllMovesForAnchor(playerState *PlayerState, anchor Position, orientation Orientation) PartialMoves {
	out := make(PartialMoves, 0, 100)
	game := state.game
	options := game.options
	corpus := game.corpus
	playerState.rack.Verify(corpus)
	fmt := game.fmt
	boardTiles := state.tileBoard
	prefixDirection := orientation.PrefixDirection()
	suffixDirection := orientation.SuffixDirection()
	ok, preceedingnPosition := state.AdjacentPosition(anchor, prefixDirection)
	if options.debug > 0 {
		fmt.Printf("GenerateAllMovesForAnchor anchor: %s orientation: %s \nplayer: %s\nanchor tile: %s\n",
			anchor.String(), orientation.String(), playerState.String(corpus),
			boardTiles[anchor.row][anchor.column].String(corpus))
	}
	if ok {
		preceedingTile := boardTiles[preceedingnPosition.row][preceedingnPosition.column]
		switch preceedingTile.kind {
		case TILE_EMPTY:
			prefixTiles := state.GetEmptyNonAnchorTiles(preceedingnPosition, prefixDirection, RackSize-1)
			maxPrefixLen := Coordinate(len(prefixTiles))
			prefixes := state.GenerateAllPrefixes(anchor, prefixDirection, playerState.rack, maxPrefixLen)
			if options.debug > 0 {
				fmt.Printf("GenerateAllMovesForAnchor... anchor: %s orientation: %s player: %s\n",
					anchor.String(), orientation.String(), playerState.String(corpus))
				if options.debug > 1 {
					fmt.Print("   prefixes:\n")
					printPartialMoves(prefixes, "       ")
				}
			}
			for _, prefix := range prefixes {
				if !prefix.endPos.equal(anchor) {
					panic("endpos of generated prefix should be the anchor (GameState.GenerateAllMovesForAnchor)")
				}
				from := PartialMove{
					id:        state.NextMoveId(),
					gameState: state,
					rack:      prefix.rack,
					startPos:  prefix.startPos,
					endPos:    prefix.endPos,
					direction: suffixDirection,
					state:     prefix.state,
					tiles:     prefix.tiles,
					score:     nil,
				}
				suffixMoves := state.GenerateAllSuffixMoves(&from)
				out = slices.Concat(out, suffixMoves)
			}

		case TILE_JOKER, TILE_LETTER:
			prefix := state.GetNonEmptyBoardTiles(preceedingnPosition, prefixDirection)
			prefixWord := game.TilesToWord(prefix)
			dawgState := game.dawg.FindPrefix(prefixWord)
			if !prefixWord.equal(dawgState.Word()) {
				msg := fmt.Sprintf("word on board %s %s not matched by dawg?? \"%s\" (GameState.GenerateAllMovesForAnchor)",
					preceedingnPosition.String(), prefixDirection.String(), prefixWord.String(game.corpus))
				panic(msg)
			}
			ok, prefixPos := state.RelativePosition(anchor, prefixDirection, Coordinate(len(prefix)))
			if !ok {
				panic(fmt.Sprintf("prefix \"%s\" from anchor %s has no valid start position (GameState.GenerateAllMovesForAnchor)", prefixWord.String(game.corpus), anchor))
			}
			from := PartialMove{
				id:        state.NextMoveId(),
				gameState: state,
				rack:      playerState.rack,
				startPos:  prefixPos,
				endPos:    anchor,
				direction: suffixDirection,
				state:     dawgState,
				tiles:     make(MoveTiles, len(prefix)),
				score:     nil,
			}
			ok, p := state.RelativePosition(anchor, prefixDirection, Coordinate(len(prefix)))
			for i, t := range prefix {
				if !ok {
					panic("could not get move position (GenerateAllMovesForAnchor)")
				}
				from.tiles[i] = MoveTile{Tile: t, pos: p, placedInMove: false}
				ok, p = state.RelativePosition(p, suffixDirection, 1)
			}
			suffixMoves := state.GenerateAllSuffixMoves(&from)
			out = slices.Concat(out, suffixMoves)

		}

	} else {
		// anchor is first tile in row/col
		// not possible to generate a prefix
		suffixMoves := state.GenerateAllSuffixMoves(&PartialMove{
			id:        state.NextMoveId(),
			gameState: state,
			rack:      playerState.rack,
			startPos:  anchor,
			endPos:    anchor,
			direction: suffixDirection,
			state:     game.dawg.initialState,
			tiles:     MoveTiles{},
			score:     nil,
		})
		out = slices.Concat(out, suffixMoves)
	}
	return out
}

func (state *GameState) GenerateAllPrefixes(anchor Position, direction Direction, rack Rack, maxLength Coordinate) PartialMoves {
	options := state.game.options
	corpus := state.game.corpus
	out := make(PartialMoves, 0, 100)
	rack.Verify(corpus)
	// first emit the zero length prefix
	pm := &PartialMove{
		id:        state.NextMoveId(),
		gameState: state,
		rack:      rack,
		startPos:  anchor,
		endPos:    anchor,
		direction: direction,
		state:     state.game.dawg.initialState,
		tiles:     make(MoveTiles, 0),
		score:     nil,
	}
	if options.debug > 0 {
		fmt.Printf("GenerateAllPrefixes emit empty prefix   anchor: %s direction: %s rack: %s maxLen: %v\n",
			anchor.String(),
			direction.String(),
			rack.String(corpus),
			maxLength)
		fmt.Printf("anchor tile: %s\n", state.tileBoard[anchor.row][anchor.column].String(corpus))

		printPartialMove(pm)
	}
	out = append(out, pm)

	// now create prefixes of length [1..maxLen]
	for prefixLength := Coordinate(1); prefixLength <= maxLength; prefixLength++ {
		ok, startPos := state.RelativePosition(anchor, direction, prefixLength)
		if !ok {
			panic(fmt.Sprintf("could not locate prefix relative position %v %s (GameState.GenerateAllPrefixes)", anchor.String(), direction.String()))
		}
		from := &PartialMove{
			id:        state.NextMoveId(),
			gameState: state,
			rack:      rack,
			startPos:  startPos,
			endPos:    startPos,
			direction: direction.Reverse(),
			state:     state.game.dawg.initialState,
			tiles:     make(MoveTiles, 0),
			score:     nil,
		}
		if options.debug > 0 {
			fmt.Printf("GenerateAllPrefixes extend prefix to %v max length: %v anchor: %s direction: %s rack: %s\n",
				prefixLength,
				maxLength,
				anchor.String(),
				direction.String(),
				rack.String(corpus))
			printPartialMove(from)
		}
		prefixes := state.GeneratePrefixes(from, prefixLength)
		out = slices.Concat(out, prefixes)
	}

	return out
}

func (state *GameState) GeneratePrefixes(from *PartialMove, length Coordinate) PartialMoves {
	out := make(PartialMoves, 0, 100)
	options := state.game.options
	if length < 1 {
		if options.debug > 0 {
			prefixLength := from.startPos.Distance(from.endPos, from.direction)
			fmt.Printf("GeneratePrefixes emit prefix length: %v\n", prefixLength)
			printPartialMove(from)
		}
		out = append(out, from)
		return out
	}
	dawg := state.game.dawg
	rackTiles := state.GenerateAllRackTiles(from.rack)
	for _, rackTile := range rackTiles {
		if !state.ValidCrossLetter(from.endPos, from.direction.Orientation(), rackTile.tile.letter) {
			continue
		}
		dawgState := dawg.Transition(from.state, rackTile.tile.letter)
		if dawgState.startNode != nil {
			_, endPos := state.AdjacentPosition(from.endPos, from.direction)
			to := &PartialMove{
				id:        state.NextMoveId(),
				gameState: state,
				rack:      rackTile.rack,
				startPos:  from.startPos,
				endPos:    endPos,
				direction: from.direction,
				state:     dawgState,
				tiles:     make(MoveTiles, len(from.tiles)+1),
				score:     nil,
			}
			copy(to.tiles, from.tiles)
			if !state.fromState.game.IsValidPos(from.endPos) {
				panic(fmt.Sprintf("expected valid from.endPos %s (GeneratePrefixes)", from.endPos.String()))
			}
			to.tiles[len(from.tiles)] = MoveTile{Tile: rackTile.tile, pos: from.endPos, placedInMove: true}

			prefixCompletion := state.GeneratePrefixes(to, length-1)
			out = slices.Concat(out, prefixCompletion)
		}
	}

	return out

}

func (state *GameState) GenerateAllRackTiles(rack Rack) RackTiles {
	out := make(RackTiles, 0, 10)
	corpus := state.game.corpus
	for i, tile := range rack {
		if slices.ContainsFunc(rack[:i], func(t Tile) bool { return t.equal(tile) }) {
			continue
		}
		if tile.kind == TILE_JOKER {
			for letter := corpus.firstLetter; letter <= corpus.lastLetter; letter++ {
				newRack := make(Rack, len(rack)-1)
				copy(newRack, rack[:i])
				copy(newRack[i:], rack[i+1:])
				r := RackTile{
					tile: Tile{kind: TILE_JOKER, letter: letter},
					rack: newRack,
				}
				r.rack.Verify(corpus)
				out = append(out, r)
			}
		} else {
			newRack := make(Rack, len(rack)-1)
			copy(newRack, rack[:i])
			copy(newRack[i:], rack[i+1:])
			r := RackTile{
				tile: Tile{kind: TILE_LETTER, letter: tile.letter},
				rack: newRack,
			}
			r.rack.Verify(corpus)
			out = append(out, r)
		}
	}
	for _, r := range out {
		r.rack.Verify(corpus)
	}
	return out
}

func (state *GameState) GenerateAllSuffixMoves(from *PartialMove) PartialMoves {
	options := state.game.options
	out := make(PartialMoves, 0, 10)
	dawg := state.game.dawg
	pos := from.endPos

	if options.debug > 0 {
		fmt.Print("GenerateAllSuffixMoves from:\n")
		printPartialMove(from)
	}

	if !state.game.IsValidPos(pos) {
		return out
	}

	if state.IsTileEmpty(pos) {
		rackTiles := state.GenerateAllRackTiles(from.rack)
		for _, rackTile := range rackTiles {
			if !state.ValidCrossLetter(pos, from.direction.Orientation(), rackTile.tile.letter) {
				continue
			}
			toState := dawg.Transition(from.state, rackTile.tile.letter)
			if toState.startNode != nil {
				_, endPos := state.AdjacentPosition(pos, from.direction)
				v := toState.LastVertex()
				to := &PartialMove{
					id:        state.NextMoveId(),
					gameState: state,
					rack:      rackTile.rack,
					startPos:  from.startPos,
					endPos:    endPos,
					direction: from.direction,
					state:     toState,
					tiles:     make(MoveTiles, len(from.tiles)+1),
					score:     nil,
				}
				copy(to.tiles, from.tiles)
				to.tiles[len(from.tiles)] = MoveTile{Tile: rackTile.tile, pos: from.endPos, placedInMove: true}
				if v.final {
					if !state.game.IsValidPos(to.endPos) || state.IsTileEmpty(to.endPos) {
						if options.debug > 0 {
							fmt.Printf("GenerateAllSuffixMoves emit\n")
							printPartialMove(to)
						}
						out = append(out, to)
					}
				}
				suffixMoves := state.GenerateAllSuffixMoves(to)
				out = slices.Concat(out, suffixMoves)
			}
		}
	} else {
		// non-empty next tile
		// proceed with the tile on the board in suffix generation
		tile := state.tileBoard[pos.row][pos.column].Tile
		toState := dawg.Transition(from.state, tile.letter)
		if toState.startNode != nil {
			_, endPos := state.AdjacentPosition(pos, from.direction)
			v := toState.LastVertex()
			to := &PartialMove{
				id:        state.NextMoveId(),
				gameState: state,
				rack:      from.rack,
				startPos:  from.startPos,
				endPos:    endPos,
				direction: from.direction,
				state:     toState,
				tiles:     make(MoveTiles, len(from.tiles)+1),
				score:     nil,
			}
			copy(to.tiles, from.tiles)
			to.tiles[len(from.tiles)] = MoveTile{Tile: tile, pos: pos, placedInMove: false}
			if v.final {
				if !state.game.IsValidPos(to.endPos) || state.IsTileEmpty(to.endPos) {
					if options.debug > 0 {
						fmt.Printf("GenerateAllSuffixMoves emit\n")
						printPartialMove(to)
					}
					out = append(out, to)
				}
			}
			suffixMoves := state.GenerateAllSuffixMoves(to)
			out = slices.Concat(out, suffixMoves)
		}
	}
	return out
}

func (state *GameState) FilterBestMove(allMoves PartialMoves) PartialMoves {
	options := state.game.options
	out := make(PartialMoves, 0, 1)
	var bestMove *PartialMove = nil

	for _, move := range allMoves {
		if move.score == nil {
			move.score = state.CalcScore(move.tiles, move.direction.Orientation())
		}
		if bestMove == nil || move.score.score > bestMove.score.score {
			if options.debug > 0 {
				fmt.Printf("\n\n################# FilterBestMove #################\n")
				curScore := Score(0)
				if bestMove != nil {
					curScore = bestMove.score.score
				}
				fmt.Printf("score %v -> %v\n", curScore, move.score.score)
				printPartialMove(move)
				fmt.Printf("\n\n")
			}
			bestMove = move
		}

	}
	if options.debug > 0 {
		fmt.Printf("\n\nEND FilterBestMove\n")
		if bestMove != nil {
			printPartialMove(bestMove, "")
		} else {
			fmt.Print("NO moves found")
		}
	}
	if bestMove != nil {
		out = append(out, bestMove)
	}
	return out
}

/*^
type RackTile struct {
	tile Tile
	rack Rack
}


type PartialMove struct {
	gameState *GameState
	rack      Rack
	startPos  Position
	endPos    Position
	direction Direction
	state     DawgState
	tiles     Tiles
	score     Score
}
*/

func (rackTile *RackTile) String(corpus *Corpus) string {
	var sb strings.Builder
	sb.WriteString("Tile: ")
	sb.WriteString(rackTile.tile.String(corpus))
	sb.WriteString(" Rack: ")
	sb.WriteString(rackTile.rack.String(corpus))
	return sb.String()
}
