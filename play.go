package main

import (
	"fmt"
	"slices"
)

type RackTile struct {
	tile Tile
	rack Rack
}

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
		printState(options.out, state)
	}
	state.PrepareMove()

	filteredPartialMoves := state.FilterBestMove(state.GenerateAllMoves(playerState))

	possiblePartialMoves := make([]*PartialMove, 0)
	for {
		m, ok := <-filteredPartialMoves
		if !ok {
			break
		}
		possiblePartialMoves = append(possiblePartialMoves, m)
	}
	if len(possiblePartialMoves) == 0 {
		return nil
	}

	newState := state.AddMove(possiblePartialMoves[0], playerState)
	return newState
}

func (state *GameState) AddMove(partial *PartialMove, playerState PlayerState) *GameState {
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
			panic(fmt.Sprintf("move generation will add new tile at %s which is not empty (GameState.AddMove)", pos.String()))
		case TILE_NONE:
			panic(fmt.Sprintf("move generation will add new tile at non-existing next position after %s direction %s (GameState.AddMove)", nextPos.String(), dir.String()))
		default:
			panic(fmt.Sprintf("move generation will add new tile of unknown kind %v (GameState.AddMove)", tile.kind))
		}
		newState.tiles[pos.row][pos.column] = BoardTile{Tile: tile, validCrossLetters: NullValidCrossLetters}
		ok, nextPos = state.AdjacentPosition(pos, dir)
		if !ok {
			panic(fmt.Sprintf("move generation will add new tile at non-existing next position after %s direction %s (GameState.AddMove)", nextPos.String(), dir.String()))
		}
		pos = nextPos
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

func (state *GameState) IsAnchor(pos Position) bool {
	return state.AnyAdjacentNonEmptyTile(pos) || (state.game.board.squares[pos.row][pos.column] == CE)
}

func (state *GameState) AnyAdjacentNonEmptyTile(pos Position) bool {
	for _, d := range AllDirections {
		t, _ := state.AdjacentTile(pos, d)
		switch t.kind {
		case TILE_JOKER, TILE_LETTER:
			return true
		}
	}
	return false
}

func (state *GameState) AdjacentTile(pos Position, d Direction) (BoardTile, Position) {
	ok, adjacentPos := state.AdjacentPosition(pos, d)

	if ok {
		return state.tiles[adjacentPos.row][adjacentPos.column], adjacentPos
	}

	return NullBoardTile, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) AdjacentPosition(pos Position, dir Direction) (bool, Position) {
	return state.RelativePosition(pos, dir, 1)

}

func (state *GameState) RelativeTile(pos Position, dir Direction, n Coordinate) (BoardTile, Position) {
	ok, relativePos := state.RelativePosition(pos, dir, n)

	if ok {
		return state.tiles[relativePos.row][relativePos.column], relativePos
	}

	return NullBoardTile, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) RelativePosition(pos Position, dir Direction, n Coordinate) (bool, Position) {
	switch dir {
	case NORTH:
		if pos.row >= n {
			return true, Position{pos.row - n, pos.column}
		}
	case SOUTH:
		if pos.row+n < state.game.height {
			return true, Position{pos.row + n, pos.column}

		}
	case WEST:
		if pos.column >= n {
			return true, Position{pos.row, pos.column - n}
		}

	case EAST:
		if pos.column+n < state.game.width {
			return true, Position{pos.row, pos.column + n}

		}
	}
	return false, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) PrepareMove() {
	game := state.game
	h := game.height
	w := game.width
	tiles := state.tiles
	for _, p := range AllOrientations {
		for r := Coordinate(0); r < h; r++ {
			for c := Coordinate(0); c < w; c++ {
				validCrossLetters := &tiles[r][c].validCrossLetters[p]
				if !validCrossLetters.ok {
					validCrossLetters.letters = state.CalcValidCrossLetters(Position{r, c}, p)
					validCrossLetters.ok = true
				}
			}
		}
	}

}

func (state *GameState) CalcValidCrossLetters(pos Position, orienttation Orientation) LetterSet {
	dawg := state.game.dawg
	validLetters := NullLetterSet
	crossDirection := orienttation.Perpendicular()
	crossPrefixDirection := crossDirection.PrefixDirection()
	crossSufixDirection := crossDirection.SuffixDirection()
	prefix := state.FindPrefix(pos, crossPrefixDirection)
	prefixEndNode := prefix.LastNode()
	suffixWord := Word{}
	ok, p := state.AdjacentPosition(pos, crossSufixDirection)
	if ok {
		suffixWord = state.GetWord(p, crossSufixDirection)
		if len(suffixWord) > 0 {
			for _, v := range prefixEndNode.vertices {
				suffix := dawg.Transitions(DawgState{startNode: v.destination, vertices: Vertices{}, word: Word{}}, suffixWord)
				if suffix.startNode != nil {
					if suffix.LastVertex().final {
						validLetters.set(v.letter)
					}
				}
			}
		} else {
			for _, v := range prefixEndNode.vertices {
				if v.final {
					validLetters.set(v.letter)
				}
			}

		}
	} else {
		if len(prefix.word) == 0 && len(suffixWord) == 0 {
			return state.game.corpus.allLetters
		}
	}
	return validLetters
}

func (state *GameState) FindPrefix(pos Position, dir Direction) DawgState {
	dawg := state.game.dawg
	tiles := state.tiles
	prefixPos := pos
	prefix := Word{}
	for {
		ok, p := state.AdjacentPosition(prefixPos, dir)
		if !ok {
			break
		}
		tile := &tiles[pos.row][pos.column]
		if tile.kind == TILE_EMPTY {
			break
		}
		if tile.kind == TILE_NONE {
			panic(fmt.Sprintf("unexpected TILE_NONE on GameState board[%v,%v] (GameState.FindPrefix)", p.row, p.column))
		}
		prefix = append(prefix, tile.letter)
		prefixPos = p
	}
	switch dir {
	case WEST, NORTH:
		slices.Reverse(prefix)
	}
	return dawg.FindPrefix(prefix)
}

func (state *GameState) GetNonEmptyBoardTiles(pos Position, dir Direction) []Tile {
	tiles := []Tile{}
	for {
		var ok bool
		tile := state.tiles[pos.row][pos.column].Tile
		if state.IsTileEmpty(pos) {
			break
		}
		tiles = append(tiles, tile)
		ok, pos = state.AdjacentPosition(pos, dir)
		if !ok {
			break
		}
	}
	switch dir {
	case WEST, NORTH:
		slices.Reverse(tiles)
	}
	return tiles
}

func (state *GameState) GetEmptyNonAnchorTiles(pos Position, dir Direction) []Tile {
	tiles := []Tile{}
	for {
		var ok bool
		tile := state.tiles[pos.row][pos.column].Tile
		if !state.IsTileEmpty(pos) {
			break
		}
		if state.IsAnchor(pos) {
			break
		}
		tiles = append(tiles, tile)
		ok, pos = state.AdjacentPosition(pos, dir)
		if !ok {
			break
		}
	}
	return tiles
}

func (state *GameState) IsTileEmpty(pos Position) bool {
	tile := &state.tiles[pos.row][pos.column]
	switch tile.kind {
	case TILE_EMPTY:
		return true
	case TILE_JOKER, TILE_LETTER:
		return false
	case TILE_NONE:
		panic(fmt.Sprintf("unexpected TILE_NONE on GameState board[%v,%v] (GameState.IsEmptyTile)", pos.row, pos.column))
	}
	panic(fmt.Sprintf("invalid tile kind %v on GameState board[%v,%v] (GameState.GetNonEmptyBoardTiles)", tile.kind, pos.row, pos.column))
}

func (state *GameState) GetWord(pos Position, dir Direction) Word {
	tiles := state.GetNonEmptyBoardTiles(pos, dir)
	word := TilesToWord(tiles)
	return word
}

func TilesToWord(tiles []Tile) Word {
	word := make(Word, 0, len(tiles))
	for _, t := range tiles {
		switch t.kind {
		case TILE_LETTER, TILE_JOKER:
			word = append(word, t.letter)
		default:
			break
		}
	}
	return word
}

type PartialMove struct {
	rack      Rack
	startPos  Position
	endPos    Position
	direction Direction
	state     DawgState
	tiles     []Tile
	score     Score
}

func (state *GameState) GenerateAllMoves(playerState PlayerState) <-chan *PartialMove {
	out := make(chan *PartialMove, 1)
	go func() {
		options := state.game.options
		fmt := state.game.fmt
		width := state.game.width
		height := state.game.height
		for r := Coordinate(0); r < height; r++ {
			anchors := state.GetAnchors(r, HORIZONTAL)
			if options.verbose {
				fmt.Fprintf(options.out, "Anchors row %d:", r)
				for _, anchor := range anchors {
					fmt.Fprintf(options.out, " (%v,%v)", anchor.row, anchor.column)
				}
				fmt.Print("\n")
			}
			for _, anchor := range anchors {
				state.GenerateAllMovesForAnchor(out, playerState, anchor, HORIZONTAL)
			}
		}
		for c := Coordinate(0); c < width; c++ {
			anchors := state.GetAnchors(c, VERTICAL)
			if options.verbose {
				fmt.Fprintf(options.out, "Anchors column %d:", c)
				for _, anchor := range anchors {
					fmt.Fprintf(options.out, " (%v,%v)", anchor.row, anchor.column)
				}
				fmt.Print("\n")
			}
			for _, anchor := range anchors {
				state.GenerateAllMovesForAnchor(out, playerState, anchor, VERTICAL)
			}
		}
		close(out)
	}()
	return out
}

func (state *GameState) GenerateAllMovesForAnchor(out chan *PartialMove, playerState PlayerState, anchor Position, orientation Orientation) {
	boardTiles := state.tiles
	prefixDirection := orientation.PrefixDirection()
	suffixDirection := orientation.SuffixDirection()
	ok, preceedingnPosition := state.AdjacentPosition(anchor, prefixDirection)
	if ok {
		preceedingTile := boardTiles[preceedingnPosition.row][preceedingnPosition.column]
		switch preceedingTile.kind {
		case TILE_EMPTY:
			prefixTiles := state.GetEmptyNonAnchorTiles(preceedingnPosition, prefixDirection)
			maxPrefixLen := Coordinate(len(prefixTiles))
			prefixes := state.GenerateAllPrefixes(anchor, prefixDirection, playerState.rack, maxPrefixLen)
			for {
				prefix, ok := <-prefixes
				if !ok {
					break
				}
				if !prefix.endPos.equal(anchor) {
					panic(fmt.Sprintf("endpos of generated prefix should be the anchor (GameState.GenerateAllMovesForAnchor)"))
				}
				state.GenerateAllSuffixMoves(out, &PartialMove{
					rack:      prefix.rack,
					startPos:  prefix.startPos,
					endPos:    prefix.endPos,
					direction: suffixDirection,
					state:     prefix.state,
					tiles:     prefix.tiles,
					score:     0,
				})
			}

		case TILE_JOKER, TILE_LETTER:
			prefix := state.GetNonEmptyBoardTiles(preceedingnPosition, prefixDirection)
			prefixWord := TilesToWord(prefix)
			dawgState := state.game.dawg.FindPrefix(prefixWord)
			if !prefixWord.equal(dawgState.word) {
				panic(fmt.Sprintf("word on board not matched by dawg?? \"%s\" (GameState.GenerateAllMovesForAnchor)", prefixWord))
			}
			ok, prefixPos := state.RelativePosition(anchor, prefixDirection, Coordinate(len(prefix)))
			if !ok {
				panic(fmt.Sprintf("prefix \"%s\" from anchor %s has no valid start position (GameState.GenerateAllMovesForAnchor)", prefixWord, anchor))
			}
			state.GenerateAllSuffixMoves(out, &PartialMove{
				rack:      playerState.rack,
				startPos:  prefixPos,
				endPos:    anchor,
				direction: suffixDirection,
				state:     dawgState,
				tiles:     prefix,
				score:     0,
			})

		}

	} else {
		// anchor is first tile in row/col
		// not possible to generate a prefix
		state.GenerateAllSuffixMoves(out, &PartialMove{
			rack:      playerState.rack,
			startPos:  anchor,
			endPos:    anchor,
			direction: suffixDirection,
			state:     state.game.dawg.initialState,
			tiles:     []Tile{},
			score:     0,
		})
	}
}

func (state *GameState) GenerateAllPrefixes(anchor Position, direction Direction, rack Rack, maxLength Coordinate) <-chan *PartialMove {
	out := make(chan *PartialMove, 1)
	go func() {
		// first emit the zero length prefix
		out <- &PartialMove{
			rack:      rack,
			startPos:  anchor,
			endPos:    anchor,
			direction: direction,
			state:     state.game.dawg.initialState,
			tiles:     make([]Tile, 0),
			score:     0,
		}

		// now create prefixes of length [1..maxLen]
		for prefixLength := Coordinate(1); prefixLength <= maxLength; prefixLength++ {
			ok, startPos := state.RelativePosition(anchor, direction, prefixLength)
			if !ok {
				panic(fmt.Sprintf("could not locate prefix relative position %v %s (GameState.GenerateAllPrefixes)", anchor.String(), direction.String()))
			}
			from := &PartialMove{
				rack:      rack,
				startPos:  startPos,
				endPos:    anchor,
				direction: direction,
				state:     state.game.dawg.initialState,
				tiles:     make([]Tile, 0),
				score:     0,
			}
			state.GeneratePrefixes(out, from, prefixLength)
		}
		close(out)
	}()
	return out
}

func (state *GameState) GeneratePrefixes(out chan *PartialMove, from *PartialMove, length Coordinate) {
	if length < 1 {
		return
	}
	dawg := state.game.dawg
	dawgState := from.state
	for rackTile := range state.GenerateAllRackTiles(from.rack) {
		dawgState = dawg.Transition(dawgState, rackTile.tile.letter)
		if dawgState.startNode != nil {
			v := dawgState.LastVertex()
			_, endPos := state.AdjacentPosition(from.endPos, from.direction)
			to := &PartialMove{
				rack:      rackTile.rack,
				startPos:  from.startPos,
				endPos:    endPos,
				direction: from.direction,
				state:     state.game.dawg.initialState,
				tiles:     append(slices.Clone(from.tiles), rackTile.tile),
				score:     0,
			}
			if v.final {
				out <- to
			}
			state.GeneratePrefixes(out, to, length-1)
		}
	}
}

func (state *GameState) GenerateAllRackTiles(rack Rack) <-chan *RackTile {
	out := make(chan *RackTile, 1)
	corpus := state.game.corpus
	go func() {
		for i, tile := range rack {
			if tile.kind == TILE_JOKER {
				for letter := corpus.firstLetter; letter <= corpus.lastLetter; letter++ {
					newRack := make(Rack, len(rack)-1)
					copy(newRack[i:], rack[i+1:])
					out <- &RackTile{
						tile: Tile{kind: TILE_JOKER, letter: letter},
						rack: newRack,
					}
				}
			} else {
				newRack := make(Rack, len(rack)-1)
				copy(newRack, rack[:i])
				copy(newRack[i:], rack[i+1:])
				out <- &RackTile{
					tile: Tile{kind: TILE_JOKER, letter: tile.letter},
					rack: newRack,
				}
			}
		}
	}()
	return out
}

func (state *GameState) GenerateAllSuffixMoves(out chan *PartialMove, from *PartialMove) {
	dawg := state.game.dawg
	rack := from.rack
	pos := from.endPos
	for len(rack) > 0 {
		if state.IsTileEmpty(pos) {
			for rackTile := range state.GenerateAllRackTiles(from.rack) {
				toState := dawg.Transition(from.state, rackTile.tile.letter)
				if toState.startNode != nil {
					_, endPos := state.AdjacentPosition(pos, from.direction)
					v := toState.LastVertex()
					to := &PartialMove{
						rack:      rackTile.rack,
						startPos:  from.startPos,
						endPos:    endPos,
						direction: from.direction,
						state:     toState,
						tiles:     append(slices.Clone(from.tiles), rackTile.tile),
						score:     0,
					}
					if v.final {
						out <- to
					}
					state.GenerateAllSuffixMoves(out, to)
				}
			}
		}
	}
}

func (state *GameState) FilterBestMove(allMoves <-chan *PartialMove) <-chan *PartialMove {
	out := make(chan *PartialMove, 1)
	go func() {
		var bestMove *PartialMove = nil
		for {
			move, ok := <-allMoves
			if !ok {
				break
			}
			if move != nil {
				if move.score > bestMove.score {
					bestMove = move
				}
			}
		}
	}()
	return out
}
