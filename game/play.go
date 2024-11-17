package game

import (
	"fmt"
	"os"
	"slices"
	"strings"
	. "wordfeud/corpus"
	. "wordfeud/dawg"
	. "wordfeud/localize"
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

func (game *_Game) Play() bool {
	options := game.options
	lang := game.options.Language

	curState := game.state
	playerNo := curState.NextPlayer()
	curPlayerStates := curState.playerStates
	curPlayerState := curPlayerStates[playerNo]
	result := false

	if curState.move == nil && options.Move > 0 {
		options.MoveDebug = options.Debug
		options.Debug = 0
	}

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
		messages := make([]string, 0)
		result = true

		for _, ps := range game.state.playerStates {
			state.FillRack(ps)
		}

		if options.Debug > 0 {
			game.fmt.Printf("game play completed move : %s\n", playerState.String(game.corpus))
		}
		for _, ps := range game.state.playerStates {
			if ps.playerNo != NoPlayer && ps.NumberOfRackTiles() == 0 {
				messages = append(messages, fmt.Sprintf(Localized(lang, "Game completed after %d moves as %s has no more tiles in rack"), state.move.seqno, ps.player.name))
				result = false
				break
			}
		}
		if result && state.consequtivePasses >= MaxConsequtivePasses {
			messages = append(messages, fmt.Sprintf(Localized(lang, "Game completed after %d moves as there has been %d conequtive passes"), state.move.seqno, state.consequtivePasses))
			result = false
		}

		if !result {
			messages = slices.Concat(messages, game.ResultMessages())
		}

		if options.WriteFile {
			gameFileName, err := WriteGameFile(game, messages)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing game file \"%s\"\n%v\n", gameFileName, err.Error())
				return false
			}
			if options.Verbose {
				messages = append(messages,
					fmt.Sprintf(Localized(lang, "Wrote game file after move %d \"%s\""),
						game.nextMoveSeqNo-1, gameFileName))
			} else {
				if !result {
					messages = append(messages, fmt.Sprintf(Localized(lang, "Game file is %s"), gameFileName))
				}
			}
		}

		if !result && len(messages) > 0 {
			fmt.Println("")
		}
		for _, m := range messages {
			fmt.Println(m)
		}

		if options.Move > 0 && move.seqno >= options.Move {
			options.Debug = options.MoveDebug
			options.Move = 0
		}
	}
	return result
}

func (game *_Game) ResultMessages() []string {
	lang := game.corpus.Language()
	messages := make([]string, 0)
	bestScore := Score(0)
	allPlayers := make(PlayerStates, 0, len(game.state.playerStates))

	for _, ps := range game.state.playerStates {
		if ps.playerNo != NoPlayer {
			allPlayers = append(allPlayers, ps)
			if ps.score > bestScore {
				bestScore = ps.score
			}
		}
	}
	slices.SortFunc(allPlayers,
		func(psl *PlayerState, psr *PlayerState) int {
			return int(psr.score) - int(psl.score)
		})
	bestScorePlayerNames := make([]string, 0, len(allPlayers))
	for _, ps := range allPlayers {
		if ps.score < bestScore {
			break
		}
		bestScorePlayerNames = append(bestScorePlayerNames, ps.player.name)
	}
	if len(bestScorePlayerNames) > 1 {
		messages = append(messages, fmt.Sprintf(Localized(lang, "Game is a draw between %d players: %s"),
			len(bestScorePlayerNames), strings.Join(bestScorePlayerNames, ", ")))
	} else {
		messages = append(messages, fmt.Sprintf(Localized(lang, "Game is won by %s"),
			bestScorePlayerNames[0]))
	}
	for _, ps := range allPlayers {
		messages = append(messages, fmt.Sprintf(Localized(lang, "%s scored %d and has %s left"),
			ps.player.name, ps.score, ps.rack.Pretty(game.corpus)))
	}
	return messages
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
	w := state.game.dimensions.Width
	h := state.game.dimensions.Height
	switch orientation {
	case HORIZONTAL:
		for c := Coordinate(0); c < w; c++ {
			switch state.tileBoard[coordinate][c].kind {
			case TILE_EMPTY:
				pos := Position{row: coordinate, column: c}
				if state.IsAnchor(pos) {
					anchors = append(anchors, pos)
				}
			}
		}

	case VERTICAL:
		for r := Coordinate(0); r < h; r++ {
			switch state.tileBoard[r][coordinate].kind {
			case TILE_EMPTY:
				pos := Position{row: r, column: coordinate}
				if state.IsAnchor(pos) {
					anchors = append(anchors, pos)
				}
			}
		}
	}
	if options.Debug > 0 {
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
	h := game.Dimensions().Height
	w := game.Dimensions().Width
	for r := Coordinate(0); r < h; r++ {
		for c := Coordinate(0); c < w; c++ {
			for _, orientation := range AllOrientations {
				validCrossLetters := &tiles[r][c].validCrossLetters[orientation]
				if !validCrossLetters.ok {
					validLetters := state.CalcValidCrossLetters(Position{r, c}, orientation)
					if game.options.Debug > 0 {
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
	if options.Debug > 0 {
		fmt.Printf("\n\n--------------------------------\n GenrateAllMoves: player: %s\n", playerState.String(corpus))
		PrintState(state)
	}
	playerState.rack.Verify(corpus)
	out := make(PartialMoves, 0, 100)
	fmt := state.game.fmt
	width := state.game.Dimensions().Width
	height := state.game.Dimensions().Height
	for r := Coordinate(0); r < height; r++ {
		anchors := state.GetAnchors(r, HORIZONTAL)
		if options.Debug > 0 {
			if len(anchors) > 0 {
				fmt.Fprintf(options.Out, "Anchors row %v: %s\n", r, anchors.String())
			}
		}
		for _, anchor := range anchors {
			moves := state.GenerateAllMovesForAnchor(playerState, anchor, HORIZONTAL)
			out = slices.Concat(out, moves)
		}
	}
	for c := Coordinate(0); c < width; c++ {
		anchors := state.GetAnchors(c, VERTICAL)
		if options.Debug > 0 {
			if len(anchors) > 0 {
				fmt.Fprintf(options.Out, "Anchors column %d %s\n", c, anchors.String())
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
	if options.Debug > 0 {
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

			if options.Debug > 0 {
				fmt.Printf("GenerateAllMovesForAnchor... anchor: %s orientation: %s #prefixes:%d player: %s\n",
					anchor.String(), orientation.String(), len(prefixes), playerState.String(corpus))
				if options.Debug > 1 {
					fmt.Print("   prefixes:\n")
					PrintPartialMoves(prefixes, "       ")
				}
			}

			for _, prefix := range prefixes {
				if options.Debug > 0 {
					fmt.Print("GenerateAllMovesForAnchor... prefix: \n")
					PrintPartialMove(prefix)
				}
				prefix.Verify()
				if !prefix.endPos.equal(anchor) {
					panic("endpos of generated prefix should be the anchor (GameState.GenerateAllMovesForAnchor)")
				}
				from := &PartialMove{
					id:        state.NextMoveId(),
					gameState: state,
					rack:      prefix.rack,
					startPos:  prefix.startPos,
					endPos:    prefix.endPos,
					direction: suffixDirection,
					state:     prefix.state,
					tiles:     make(MoveTiles, len(prefix.tiles)),
					score:     nil,
				}

				ok, p := state.RelativePosition(anchor, prefixDirection, Coordinate(len(prefix.tiles)))
				for i, t := range prefix.tiles {
					if !ok {
						panic("could not get move position (GenerateAllMovesForAnchor)")
					}
					from.tiles[i] = MoveTile{Tile: t.Tile, pos: p, placedInMove: true}
					ok, p = state.RelativePosition(p, suffixDirection, 1)
				}
				from.Verify()
				suffixMoves := state.GenerateAllSuffixMoves(from)
				out = slices.Concat(out, suffixMoves)
			}

		case TILE_JOKER, TILE_LETTER:
			prefix := state.GetNonEmptyBoardTiles(preceedingnPosition, prefixDirection)
			prefixWord := game.TilesToWord(prefix)
			dawgState := game.dawg.FindPrefix(prefixWord)
			if !prefixWord.Equal(dawgState.Word()) {
				msg := fmt.Sprintf("word on board %s %s not matched by dawg?? \"%s\" (GameState.GenerateAllMovesForAnchor)",
					preceedingnPosition.String(), prefixDirection.String(), prefixWord.String(game.corpus))
				panic(msg)
			}
			ok, prefixPos := state.RelativePosition(anchor, prefixDirection, Coordinate(len(prefix)))
			if !ok {
				panic(fmt.Sprintf("prefix \"%s\" from anchor %s has no valid start position (GameState.GenerateAllMovesForAnchor)", prefixWord.String(game.corpus), anchor))
			}
			from := &PartialMove{
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
			from.Verify()
			suffixMoves := state.GenerateAllSuffixMoves(from)
			out = slices.Concat(out, suffixMoves)

		}

	} else {
		// anchor is first tile in row/col
		// not possible to generate a prefix
		from := &PartialMove{
			id:        state.NextMoveId(),
			gameState: state,
			rack:      playerState.rack,
			startPos:  anchor,
			endPos:    anchor,
			direction: suffixDirection,
			state:     game.dawg.InitialState(),
			tiles:     MoveTiles{},
			score:     nil,
		}
		from.Verify()
		suffixMoves := state.GenerateAllSuffixMoves(from)
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
		state:     state.game.dawg.InitialState(),
		tiles:     make(MoveTiles, 0),
		score:     nil,
	}
	if options.Debug > 0 {
		fmt.Printf("GenerateAllPrefixes emit empty prefix   anchor: %s direction: %s rack: %s maxLen: %v\n",
			anchor.String(),
			direction.String(),
			rack.String(corpus),
			maxLength)
		fmt.Printf("anchor tile: %s\n", state.tileBoard[anchor.row][anchor.column].String(corpus))
		PrintPartialMove(pm)
	}
	pm.Verify()
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
			state:     state.game.dawg.InitialState(),
			tiles:     make(MoveTiles, 0),
			score:     nil,
		}
		if options.Debug > 0 {
			fmt.Printf("GenerateAllPrefixes extend prefix to %v max length: %v anchor: %s direction: %s rack: %s\n",
				prefixLength,
				maxLength,
				anchor.String(),
				direction.String(),
				rack.String(corpus))
			PrintPartialMove(from)
		}
		from.Verify()
		prefixes := state.GeneratePrefixes(from, prefixLength)
		out = slices.Concat(out, prefixes)
	}

	return out
}

func (state *GameState) GeneratePrefixes(from *PartialMove, length Coordinate) PartialMoves {
	out := make(PartialMoves, 0, 100)
	options := state.game.options
	if length < 1 {
		if options.Debug > 0 {
			prefixLength := from.startPos.Distance(from.endPos, from.direction)
			fmt.Printf("GeneratePrefixes emit prefix length: %v\n", prefixLength)
			PrintPartialMove(from)
		}
		from.Verify()
		out = append(out, from)
		return out
	}
	rackTiles := state.GenerateAllRackTiles(from.rack)
	for _, rackTile := range rackTiles {
		if !state.ValidCrossLetter(from.endPos, from.direction.Orientation(), rackTile.tile.letter) {
			continue
		}
		dawgState := from.state.Transition(rackTile.tile.letter)
		if dawgState.Valid() {
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
			to.Verify()
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
			for letter, last := corpus.FirstLetter(), corpus.LastLetter(); letter <= last; letter++ {
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
	pos := from.endPos

	if options.Debug > 0 {
		fmt.Print("GenerateAllSuffixMoves from:\n")
		PrintPartialMove(from)
	}
	from.Verify()
	if !state.game.IsValidPos(pos) {
		return out
	}

	if state.IsTileEmpty(pos) {
		rackTiles := state.GenerateAllRackTiles(from.rack)
		for _, rackTile := range rackTiles {
			if !state.ValidCrossLetter(pos, from.direction.Orientation(), rackTile.tile.letter) {
				continue
			}
			toState := from.state.Transition(rackTile.tile.letter)
			if toState.Valid() {
				_, endPos := state.AdjacentPosition(pos, from.direction)
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
				to.Verify()
				if toState.Final() {
					if !state.game.IsValidPos(to.endPos) || state.IsTileEmpty(to.endPos) {
						if options.Debug > 0 {
							fmt.Printf("GenerateAllSuffixMoves emit\n")
							PrintPartialMove(to)
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
		toState := from.state.Transition(tile.letter)
		if toState.Valid() {
			_, endPos := state.AdjacentPosition(pos, from.direction)
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
			to.Verify()
			if toState.Final() {
				if !state.game.IsValidPos(to.endPos) || state.IsTileEmpty(to.endPos) {
					if options.Debug > 0 {
						fmt.Printf("GenerateAllSuffixMoves emit\n")
						PrintPartialMove(to)
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
			if options.Debug > 0 {
				fmt.Printf("\n\n################# FilterBestMove #################\n")
				curScore := Score(0)
				if bestMove != nil {
					curScore = bestMove.score.score
				}
				fmt.Printf("score %v -> %v\n", curScore, move.score.score)
				PrintPartialMove(move)
				fmt.Printf("\n\n")
				move.Verify()
			}
			bestMove = move
		}

	}
	if options.Debug > 0 {
		fmt.Printf("\n\nEND FilterBestMove\n")
		if bestMove != nil {
			PrintPartialMove(bestMove, "")
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

func (rackTile *RackTile) String(corpus Corpus) string {
	var sb strings.Builder
	sb.WriteString("Tile: ")
	sb.WriteString(rackTile.tile.String(corpus))
	sb.WriteString(" Rack: ")
	sb.WriteString(rackTile.rack.String(corpus))
	return sb.String()
}

var inPartialMoveVerify uint = 0

func (pm *PartialMove) Verify() {
	if inPartialMoveVerify > 0 {
		return
	}
	inPartialMoveVerify++
	defer func() {
		inPartialMoveVerify--
	}()
	gameState := pm.gameState
	game := gameState.game
	corpus := game.corpus
	tilesWord := gameState.TilesToString(pm.tiles.Tiles())
	dawgStateWord := pm.state.Word().String(corpus)
	if tilesWord != dawgStateWord {
		message := game.fmt.Sprintf("BAD PartialMove - tilesWord(%s) != dawgStateWord(%s) :\n", tilesWord, dawgStateWord)
		game.fmt.Print(message)
		PrintPartialMove(pm)
		panic(message)
	}
}
