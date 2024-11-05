package main

import (
	"fmt"
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
	tiles     Tiles
	score     Score
}

type PartialMoves []*PartialMove

func (game *Game) play() bool {
	state := game.state
	playerNo := state.nextPlayer()
	playerState := state.playerStates[playerNo]
	newState := state.Move(playerState)
	if newState != nil {
		game.state = newState
		return true
	}
	return false
}

func (state *GameState) nextPlayer() PlayerNo {
	if state.move == nil {
		return 0
	}
	for i, p := range state.game.players {
		if state.move.playerState.player == p {
			n := i + 1
			if n > len(state.game.players) {
				return 0
			}
			return PlayerNo(n)
		}
	}
	return 0
}

func (state *GameState) Move(playerState PlayerState) *GameState {
	fmt := state.game.fmt
	options := state.game.options
	if options.verbose {
		fmt.Fprintf(options.out, "\n\nMove for player %v : %s\n", playerState.no, playerState.player.name)
		fprintState(options.out, state)
	}
	state.PrepareMove()

	filteredPartialMoves := state.FilterBestMove(state.GenerateAllMoves(playerState))

	if len(filteredPartialMoves) == 0 {
		return nil
	}

	newState := state.AddMove(filteredPartialMoves[0], playerState)
	return newState
}

func (state *GameState) AddMove(partial *PartialMove, playerState PlayerState) *GameState {
	options := state.game.options
	playerState.rack = partial.rack
	move := state.MakeMove(partial.startPos, partial.direction, partial.tiles, playerState)
	newState := &GameState{
		game:         state.game,
		fromState:    state,
		move:         move,
		playerStates: state.playerStates,
	}
	height := state.game.height
	width := state.game.width
	newState.tiles = make([][]BoardTile, height)
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
	var ok bool
	var nextPos Position
	for _, tile := range partial.tiles {
		boardTile := newState.tiles[pos.row][pos.column]
		switch boardTile.kind {
		case TILE_EMPTY:
		case TILE_JOKER, TILE_LETTER:
			if !boardTile.Tile.equal(tile) {
				panic(fmt.Sprintf("move generation will add new tile %v at %s which is not empty and differs %v (GameState.AddMove)", tile, pos.String(), boardTile.Tile))
			}
			// this is a tile from a previous move -- skip
			continue
		case TILE_NONE:
			panic(fmt.Sprintf("move generation will add new tile at non-existing next position after %s direction %s (GameState.AddMove)", nextPos.String(), dir.String()))
		default:
			panic(fmt.Sprintf("move generation will add new tile of unknown kind %v (GameState.AddMove)", tile.kind))
		}
		newState.tiles[pos.row][pos.column] = BoardTile{Tile: tile, validCrossLetters: NullValidCrossLetters}
		switch dir.Orientation() {
		case HORIZONTAL:
		case VERTICAL:
		}
		ok, nextPos = state.AdjacentPosition(pos, dir)
		if !ok {
			panic(fmt.Sprintf("move generation will add new tile at non-existing next position after %s direction %s (GameState.AddMove)", nextPos.String(), dir.String()))
		}
		pos = nextPos
	}
	if options.debug > 0 {
		printState(state)
		fmt.Printf("AddMove :")
		printMove(move)
	}
	return newState
}

func (state *GameState) GetAnchors(coordinate Coordinate, orientation Orientation) Positions {
	anchors := make(Positions, 0)
	switch orientation {
	case HORIZONTAL:
		for c := Coordinate(0); c < state.game.width; c++ {
			switch state.tiles[coordinate][c].kind {
			case TILE_EMPTY:
				pos := Position{row: coordinate, column: c}
				if state.IsAnchor(pos) {
					anchors = append(anchors, pos)
				}
			}
		}

	case VERTICAL:
		for r := Coordinate(0); r < state.game.height; r++ {
			switch state.tiles[r][coordinate].kind {
			case TILE_EMPTY:
				pos := Position{row: r, column: coordinate}
				if state.IsAnchor(pos) {
					anchors = append(anchors, pos)
				}
			}
		}
	}
	return anchors
}

func (state *GameState) PrepareMove() {
	game := state.game
	h := game.height
	w := game.width
	tiles := state.tiles
	for _, orientation := range AllOrientations {
		for r := Coordinate(0); r < h; r++ {
			for c := Coordinate(0); c < w; c++ {
				validCrossLetters := &tiles[r][c].validCrossLetters[orientation]
				if !validCrossLetters.ok {
					validCrossLetters.letters = state.CalcValidCrossLetters(Position{r, c}, orientation)
					validCrossLetters.ok = true
				}
			}
		}
	}

}

func (state *GameState) GenerateAllMoves(playerState PlayerState) PartialMoves {
	options := state.game.options
	if options.debug > 0 {
		fmt.Print("\n\n--------------------------------\n GenrateAllMoves:\n")
		printState(state)
	}
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
				fmt.Fprintf(options.out, "Anchors column %dv %s\n", c, anchors.String())
			}
		}
		for _, anchor := range anchors {
			moves := state.GenerateAllMovesForAnchor(playerState, anchor, VERTICAL)
			out = slices.Concat(out, moves)
		}
	}
	return out
}

func (state *GameState) GenerateAllMovesForAnchor(playerState PlayerState, anchor Position, orientation Orientation) PartialMoves {
	out := make(PartialMoves, 0, 100)
	game := state.game
	options := game.options
	corpus := game.corpus
	fmt := game.fmt
	boardTiles := state.tiles
	prefixDirection := orientation.PrefixDirection()
	suffixDirection := orientation.SuffixDirection()
	ok, preceedingnPosition := state.AdjacentPosition(anchor, prefixDirection)
	if options.debug > 0 {
		fmt.Printf("GenerateAllMovesForAnchor anchor: %s orientation: %s player: %s\n",
			anchor.String(), orientation.String(), playerState.String(corpus))
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
				suffixMoves := state.GenerateAllSuffixMoves(&PartialMove{
					id:        state.NextMoveId(),
					gameState: state,
					rack:      prefix.rack,
					startPos:  prefix.startPos,
					endPos:    prefix.endPos,
					direction: suffixDirection,
					state:     prefix.state,
					tiles:     prefix.tiles,
					score:     0,
				})
				out = slices.Concat(out, suffixMoves)
			}

		case TILE_JOKER, TILE_LETTER:
			prefix := state.GetNonEmptyBoardTiles(preceedingnPosition, prefixDirection)
			prefixWord := game.TilesToWord(prefix)
			dawgState := game.dawg.FindPrefix(prefixWord)
			if !prefixWord.equal(dawgState.word) {
				panic(fmt.Sprintf("word on board not matched by dawg?? \"%s\" (GameState.GenerateAllMovesForAnchor)", prefixWord.String(game.corpus)))
			}
			ok, prefixPos := state.RelativePosition(anchor, prefixDirection, Coordinate(len(prefix)))
			if !ok {
				panic(fmt.Sprintf("prefix \"%s\" from anchor %s has no valid start position (GameState.GenerateAllMovesForAnchor)", prefixWord.String(game.corpus), anchor))
			}
			suffixMoves := state.GenerateAllSuffixMoves(&PartialMove{
				id:        state.NextMoveId(),
				gameState: state,
				rack:      playerState.rack,
				startPos:  prefixPos,
				endPos:    anchor,
				direction: suffixDirection,
				state:     dawgState,
				tiles:     prefix,
				score:     0,
			})
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
			tiles:     Tiles{},
			score:     0,
		})
		out = slices.Concat(out, suffixMoves)
	}
	return out
}

func (state *GameState) GenerateAllPrefixes(anchor Position, direction Direction, rack Rack, maxLength Coordinate) PartialMoves {
	options := state.game.options
	corpus := state.game.corpus
	out := make(PartialMoves, 0, 100)
	// first emit the zero length prefix
	pm := &PartialMove{
		id:        state.NextMoveId(),
		gameState: state,
		rack:      rack,
		startPos:  anchor,
		endPos:    anchor,
		direction: direction,
		state:     state.game.dawg.initialState,
		tiles:     make(Tiles, 0),
		score:     0,
	}
	if options.debug > 0 {
		fmt.Printf("GenerateAllPrefixes emit empty prefix   anchor: %s direction: %s rack: %s maxLen: %v\n",
			anchor.String(),
			direction.String(),
			rack.String(corpus),
			maxLength)
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
			tiles:     make(Tiles, 0),
			score:     0,
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
				tiles:     append(slices.Clone(from.tiles), rackTile.tile),
				score:     0,
			}

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
				copy(newRack[i:], rack[i+1:])
				out = append(out, RackTile{
					tile: Tile{kind: TILE_JOKER, letter: letter},
					rack: newRack,
				})
			}
		} else {
			newRack := make(Rack, len(rack)-1)
			copy(newRack, rack[:i])
			copy(newRack[i:], rack[i+1:])
			out = append(out, RackTile{
				tile: Tile{kind: TILE_LETTER, letter: tile.letter},
				rack: newRack,
			})
		}
	}
	return out
}

func (state *GameState) GenerateAllSuffixMoves(from *PartialMove) PartialMoves {
	options := state.game.options
	out := make(PartialMoves, 0, 100)
	dawg := state.game.dawg
	pos := from.endPos

	if options.debug > 0 {
		fmt.Print("GenerateAllSuffixMoves from:\n")
		printPartialMove(from)
	}

	if state.IsTileEmpty(pos) {
		rackTiles := state.GenerateAllRackTiles(from.rack)
		for _, rackTile := range rackTiles {
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
					tiles:     append(slices.Clone(from.tiles), rackTile.tile),
					score:     0,
				}
				if v.final {
					to.score = state.CalcScore(to.startPos, to.direction, to.tiles)
					if options.debug > 0 {
						fmt.Printf("GenerateAllSuffixMoves emit\n")
						printPartialMove(to)
					}
					out = append(out, to)
				}
				suffixMoves := state.GenerateAllSuffixMoves(to)
				out = slices.Concat(out, suffixMoves)
			}
		}

	}
	return out
}

func (state *GameState) FilterBestMove(allMoves PartialMoves) PartialMoves {
	options := state.game.options
	out := make(PartialMoves, 0, 1)
	var bestMove *PartialMove = nil
	for _, move := range allMoves {
		if bestMove == nil || move.score > bestMove.score {
			if options.debug > 0 {
				fmt.Printf("\n\n################# FilterBestMove #################\n")
				printPartialMove(move)
				fmt.Printf("\n\n")
			}
			bestMove = move
		}

	}
	if options.debug > 0 {
		fmt.Printf("\n\nEND FilterBestMove\n")
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
