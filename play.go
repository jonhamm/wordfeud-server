package main

import (
	"fmt"
	"slices"
)

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

func (state *GameState) Move(playerState *PlayerState) *GameState {
	p := state.game.fmt
	options := state.game.options
	if options.verbose {
		state.game.fmt.Fprintf(options.out, "\n\nMove for player %v : %s\n", playerState.no, playerState.player.name)
		printState(options.out, state)
	}
	state.PrepareMove()
	anchors := state.GetAnchors()

	if options.verbose {
		p.Fprintf(options.out, "\nAnchors:\n")
		for _, a := range anchors {
			p.Fprintf(options.out, "   [%v,%v] \n", a.row, a.column)
		}
	}

	possibleMoves := state.GenerateAllMoves(anchors)
	filteredMoves := state.FilterBestMove(possibleMoves)

	_, ok := <-filteredMoves
	if !ok {
		return nil
	}

	return nil
}

func (state *GameState) GetAnchors() Positions {
	anchors := make(Positions, 0)
	for r := Coordinate(0); r < state.game.height; r++ {
		for c := Coordinate(0); c < state.game.width; c++ {
			pos := Position{row: r, column: c}
			switch state.tiles[r][c].kind {
			case TILE_EMPTY:
				_, d, pos := state.AdjacentNonEmptyTile(pos)
				if d != NONE {
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

func (state *GameState) AdjacentNonEmptyTile(pos Position) (BoardTile, Direction, Position) {
	for _, d := range AllDirections {
		t, p := state.AdjacentTile(pos, d)
		switch t.kind {
		case TILE_JOKER, TILE_LETTER:
			return t, d, p
		}
	}
	return NullBoardTile, NONE, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) AdjacentTile(pos Position, d Direction) (BoardTile, Position) {
	ok, adjacentPos := state.AdjacentPosition(pos, d)

	if ok {
		return state.tiles[adjacentPos.row][adjacentPos.column], adjacentPos
	}

	return NullBoardTile, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) AdjacentPosition(pos Position, d Direction) (bool, Position) {
	switch d {
	case NORTH:
		if pos.row > 0 {
			return true, Position{pos.row - 1, pos.column}
		}
	case SOUTH:
		if pos.row+1 < state.game.height {
			return true, Position{pos.row + 1, pos.column}

		}
	case WEST:
		if pos.column > 0 {
			return true, Position{pos.row, pos.column}
		}

	case EAST:
		if pos.column+1 < state.game.width {
			return true, Position{pos.row, pos.column + 1}

		}
	}
	return false, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) PrepareMove() {
	game := state.game
	h := game.height
	w := game.width
	tiles := state.tiles
	for _, p := range AllPlanes {
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
	directions := orienttation.Inverse().Directions()
	prefix := state.FindPrefix(pos, directions[0])
	prefixEndNode := prefix.LastNode()
	suffixWord := Word{}
	ok, p := state.AdjacentPosition(pos, directions[1])
	if ok {
		suffixWord = state.GetWord(p, directions[1])
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

func (state *GameState) GetWord(pos Position, dir Direction) Word {
	word := Word{}
	tiles := state.tiles
	for {
		ok, pos := state.AdjacentPosition(pos, dir)
		if !ok {
			break
		}
		tile := &tiles[pos.row][pos.column]
		if tile.kind == TILE_EMPTY {
			break
		}
		if tile.kind == TILE_NONE {
			panic(fmt.Sprintf("unexpected TILE_NONE on GameState board[%v,%v] (GameState.GetWord)", pos.row, pos.column))
		}
		word = append(word, tile.letter)
	}
	return word
}

func (state *GameState) GenerateAllMoves(anchors Positions) <-chan *Move {
	out := make(chan *Move, 100)
	go func() {
		for _, anchor := range anchors {
			state.GenerateAllMovesForAnchor(out, anchor)
			/*		rack := state.move.playerState.rack
					tiles := state.tiles
					board := state.game.board */
		}
		close(out)
	}()
	return out
}

func (state *GameState) GenerateAllMovesForAnchor(out chan *Move, anchor Position) {

}

func (state *GameState) FilterBestMove(allMoves <-chan *Move) <-chan *Move {
	out := make(chan *Move, 10)
	go func() {
		var bestMove *Move = nil
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
