package main

import (
	"fmt"
	"slices"
	"strings"
	. "wordfeud/corpus"
)

type TileKind byte

const (
	TILE_EMPTY  = TileKind(0)
	TILE_JOKER  = TileKind(1)
	TILE_LETTER = TileKind(2)
	TILE_NONE   = TileKind(3)
)

type Direction byte
type Directions []Direction
type DirectionSet byte

const (
	NONE  = Direction(0)
	NORTH = Direction(1)
	SOUTH = Direction(2)
	WEST  = Direction(3)
	EAST  = Direction(4)
)

var AllDirections = Directions{NORTH, SOUTH, EAST, WEST}

type Orientation byte
type Planes []Orientation

const (
	HORIZONTAL = Orientation(0)
	VERTICAL   = Orientation(1)
)

var AllOrientations = Planes{HORIZONTAL, VERTICAL}

const PlaneMax = VERTICAL + 1

type Tile struct {
	kind   TileKind
	letter Letter
}

type Tiles []Tile

var NullTile = Tile{kind: TILE_NONE, letter: 0}

type ValidCrossLetters struct {
	ok      bool
	letters LetterSet
}

var NullValidCrossLetters = [PlaneMax]ValidCrossLetters{
	{ok: false, letters: 0},
	{ok: false, letters: 0},
}
var NoValidCrossLetters = [PlaneMax]ValidCrossLetters{
	{ok: true, letters: 0},
	{ok: true, letters: 0},
}

type BoardTile struct {
	Tile
	validCrossLetters [PlaneMax]ValidCrossLetters
}

var NullBoardTile = BoardTile{
	Tile:              NullTile,
	validCrossLetters: NullValidCrossLetters,
}

type TileBoard [][]BoardTile
type GameState struct {
	game              *Game
	fromState         *GameState
	move              *Move
	tileBoard         TileBoard
	playerStates      PlayerStates
	playerNo          PlayerNo
	freeTiles         Tiles
	consequtivePasses int
}

type GameStates []*GameState

const RackSize = 7

type Rack Tiles

type PlayerState struct {
	player   *Player
	playerNo PlayerNo
	score    Score
	rack     Rack
}

type PlayerStates []*PlayerState

func InitialGameState(game *Game) (*GameState, error) {
	options := game.options
	corpus := game.corpus
	state := &GameState{game: game, fromState: nil, move: nil, tileBoard: make(TileBoard, game.height)}
	allLetters := game.corpus.AllLetters()

	languageTiles := GetLanguageTiles(options.language)
	for _, tile := range languageTiles {
		game.letterScores[game.corpus.RuneToLetter(tile.Character())] = Score(tile.Value())
		state.freeTiles = slices.Grow(state.freeTiles, len(state.freeTiles)+int(tile.Count()))
		for i, n := byte(0), byte(tile.Count()); i < n; i++ {
			state.freeTiles = append(state.freeTiles, Tile{TILE_LETTER, corpus.RuneToLetter(tile.Character())})
		}
	}
	for i := 0; byte(i) < JOKER_COUNT; i++ {
		state.freeTiles = append(state.freeTiles, Tile{TILE_JOKER, 0})
	}

	for r := Coordinate(0); r < game.height; r++ {
		state.tileBoard[r] = make([]BoardTile, game.width)
		for c := Coordinate(0); c < game.width; c++ {
			for p := range AllOrientations {
				validCrossLetters := &state.tileBoard[r][c].validCrossLetters[p]
				validCrossLetters.ok = true
				validCrossLetters.letters = allLetters
			}
		}
	}

	if options.debug > 0 {
		game.fmt.Printf("initial free tiles: (%d) %s\n", len(state.freeTiles), state.freeTiles.String(corpus))
	}

	state.playerStates = make(PlayerStates, len(game.players))
	for i, player := range game.players {
		state.playerStates[i] = &PlayerState{
			player:   player,
			playerNo: PlayerNo(i),
			score:    0,
			rack:     Rack{},
		}
		state.FillRack(state.playerStates[i])
	}
	return state, nil
}

func (state *GameState) TakeTile() Tile {
	n := len(state.freeTiles)
	if n == 0 {
		return Tile{TILE_EMPTY, 0}
	}
	i := state.game.rand.Intn(n)
	t := state.freeTiles[i]
	state.freeTiles = slices.Delete(state.freeTiles, i, i+1)
	return t
}

func (state *GameState) FillRack(playerState *PlayerState) {
	game := state.game
	options := game.options
	fmt := game.fmt
	if options.debug > 1 {
		fmt.Printf("FillRack %s\n", playerState.String(game.corpus))
	}
	if playerState.playerNo == NoPlayer {
		return
	}
	rack := slices.Clone(playerState.rack)
	rack = slices.Grow(rack, RackSize)
	for i := len(rack); i < int(RackSize); i++ {
		t := state.TakeTile()
		if t.kind == TILE_EMPTY {
			break
		}
		rack = append(rack, t)
	}
	if options.debug > 1 {
		fmt.Printf("  => %s\n", rack.String(game.corpus))
		fmt.Printf("freeTiles (%d) %s\n", len(state.freeTiles), state.freeTiles.String(game.corpus))
	}
	for _, t := range rack {
		switch t.kind {
		case TILE_EMPTY, TILE_NONE:
			panic(fmt.Sprintf("invalid rack tile %s", t.String(state.game.corpus)))
		case TILE_JOKER, TILE_LETTER:
		}
	}
	playerState.rack = rack
}

func (state *GameState) NextMoveId() uint {
	return state.game.NextMoveId()
}

func (state *GameState) CalcValidCrossLetters(pos Position, orienttation Orientation) LetterSet {
	options := state.game.options
	fmt := state.game.fmt
	if options.debug > 0 {
		fmt.Printf("CalcValidCrossLetters %s %s tile: %s\n",
			pos.String(), orienttation.String(), state.tileBoard[pos.row][pos.column].String(state.game.corpus))
	}
	if !state.IsTileEmpty(pos) {
		if options.debug > 0 {
			fmt.Printf("CalcValidCrossLetters... non-empty tile => %s\n", state.game.corpus.AllLetters().String(state.game.corpus))
		}
		return NullLetterSet
	}
	dawg := state.game.dawg
	validLetters := NullLetterSet
	crossDirection := orienttation.Perpendicular()
	crossPrefixDirection := crossDirection.PrefixDirection()
	crossSufixDirection := crossDirection.SuffixDirection()
	prefix := state.FindPrefix(pos, crossPrefixDirection)
	prefixEndNode := prefix.LastNode()
	if options.debug > 0 {
		fmt.Printf("   prefix: %s\n", prefix.Word().String(state.game.corpus))
	}

	ok, p := state.AdjacentPosition(pos, crossSufixDirection)
	if ok {
		suffixWord := state.GetWord(p, crossSufixDirection)
		if len(suffixWord) > 0 {
			if options.debug > 0 {
				fmt.Printf("   suffix: %s\n", suffixWord.String(state.game.corpus))
			}

			for _, v := range prefixEndNode.vertices {
				suffix := dawg.Transitions(DawgState{startNode: v.destination, vertices: Vertices{}}, suffixWord)
				if suffix.startNode != nil {
					if suffix.LastVertex().final {
						validLetters.Set(v.letter)
					}
				}
			}
		} else {
			for _, v := range prefixEndNode.vertices {
				if v.final {
					validLetters.Set(v.letter)
				}
			}

		}
	} else { // no suffix
		if prefix.WordLength() == 0 {
			//neither suffix nor prefix
			if options.debug > 0 {
				fmt.Printf("CalcValidCrossLetters... isolated tilel => %s\n", state.game.corpus.AllLetters().String(state.game.corpus))
			}
			return state.game.corpus.AllLetters()
		}
		for _, v := range prefixEndNode.vertices {
			if v.final {
				validLetters.Set(v.letter)
			}
		}
	}
	if options.debug > 0 {
		fmt.Printf("CalcValidCrossLetters...  => %s\n", validLetters.String(state.game.corpus))
	}
	return validLetters
}

func (state *GameState) ValidCrossLetter(pos Position, orientation Orientation, letter Letter) bool {
	validCrossLetters := &state.tileBoard[pos.row][pos.column].validCrossLetters[orientation]
	if !validCrossLetters.ok {
		validCrossLetters.letters = state.CalcValidCrossLetters(pos, orientation)
		validCrossLetters.ok = true
	}
	return validCrossLetters.letters.Test(letter)
}

func (state *GameState) FilledPositions() Positions {
	height := state.game.height
	width := state.game.width
	filled := make(Positions, 0)
	for r := Coordinate(0); r < height; r++ {
		for c := Coordinate(0); c < width; c++ {
			p := Position{r, c}
			if !state.IsTileEmpty(p) {
				filled = append(filled, p)
			}
		}
	}
	return filled
}

func (tileBoard TileBoard) Clone() TileBoard {
	clone := make([][]BoardTile, len(tileBoard))
	for i, r := range tileBoard {
		clone[i] = slices.Clone(r)
	}
	return clone
}

func (state *GameState) NumberOfRackTiles() int {
	n := 0
	for _, p := range state.playerStates {
		n += p.NumberOfRackTiles()
	}
	return n
}

func (playerState *PlayerState) NumberOfRackTiles() int {
	return len(playerState.rack)
}

func (d Direction) Reverse() Direction {
	switch d {
	case NONE:
		return NONE
	case NORTH:
		return SOUTH
	case SOUTH:
		return NORTH
	case WEST:
		return EAST
	case EAST:
		return WEST
	}
	panic(fmt.Sprintf("invalid direction %d  (Direction.Reverse)", d))
}

func (d Direction) Orientation() Orientation {
	switch d {
	case NONE:
		panic("direction NULL has no orientation (Direction.Orientation)")
	case NORTH, SOUTH:
		return VERTICAL
	case WEST, EAST:
		return HORIZONTAL
	}
	panic(fmt.Sprintf("invalid direction %d  (Direction.Orientation)", d))
}

func (o Orientation) Directions() Directions {
	switch o {
	case HORIZONTAL:
		return Directions{WEST, EAST}
	case VERTICAL:
		return Directions{NORTH, SOUTH}
	}
	panic("invalid orientation (Orientation.Directions)")
}

func (o Orientation) PrefixDirection() Direction {
	return o.Directions()[0]
}

func (o Orientation) SuffixDirection() Direction {
	return o.Directions()[1]
}

func (o Orientation) Perpendicular() Orientation {
	switch o {
	case HORIZONTAL:
		return VERTICAL
	case VERTICAL:
		return HORIZONTAL
	}
	panic("invalid plane (Orientation.Perpendicular)")
}

func (directionSet DirectionSet) test(dir Direction) bool {
	return (directionSet & (1 << dir)) != 0
}

func (directionSet *DirectionSet) set(dir Direction) *DirectionSet {
	*directionSet |= DirectionSet(1 << dir)
	return directionSet
}

func (directionSet *DirectionSet) unset(dir Direction) *DirectionSet {
	*directionSet &^= DirectionSet(1 << dir)
	return directionSet
}

func (kind TileKind) String() string {
	switch kind {
	case TILE_EMPTY:
		return "="
	case TILE_JOKER:
		return "?"
	case TILE_LETTER:
		return "+"
	case TILE_NONE:
		return "-"
	}
	panic(fmt.Sprintf("invalid TileKind %d", kind))
}

func (orientation Orientation) String() string {
	switch orientation {
	case HORIZONTAL:
		return "horizontal"
	case VERTICAL:
		return "vertical"
	}
	panic(fmt.Sprintf("invalid orientation %d", orientation))

}

func (dir Direction) String() string {
	switch dir {
	case NONE:
		return "NONE"
	case NORTH:
		return "N"
	case SOUTH:
		return "S"
	case EAST:
		return "E"
	case WEST:
		return "W"
	}
	panic(fmt.Sprintf("invalid Direction %d", dir))
}

func (dirs Directions) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, dir := range dirs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(dir.String())
	}
	sb.WriteRune(']')
	return sb.String()
}

func (directionSet *DirectionSet) String(corpus Corpus) string {
	var s strings.Builder
	var first = true
	s.WriteRune('{')
	for _, dir := range AllDirections {
		if directionSet.test(dir) {
			if first {
				first = false
			} else {
				s.WriteRune(',')
			}
			s.WriteString(dir.String())
		}
	}
	s.WriteRune('}')
	return s.String()
}

func (player *PlayerState) String(corpus Corpus) string {
	return fmt.Sprintf("[%d] id: %v : %s score: %v rack: %v",
		player.playerNo, player.player.id, player.player.name, player.score, player.rack.String(corpus))
}

func (lhs Tile) equal(rhs Tile) bool {
	return lhs.kind == rhs.kind && lhs.letter == rhs.letter
}

func (tile Tile) String(corpus Corpus) string {
	return fmt.Sprintf("'%s':%v:%s", tile.letter.String(corpus), tile.letter, tile.kind.String())

}

func (tiles Tiles) String(corpus Corpus) string {
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

func (state *GameState) IsAnchor(pos Position) bool {
	return state.IsTileEmpty(pos) && (state.AnyAdjacentNonEmptyTile(pos) || (state.game.board.squares[pos.row][pos.column] == CE))
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
		return state.tileBoard[adjacentPos.row][adjacentPos.column], adjacentPos
	}

	return NullBoardTile, Position{state.game.height + 1, state.game.width + 1}
}

func (state *GameState) AdjacentPosition(pos Position, dir Direction) (bool, Position) {
	return state.RelativePosition(pos, dir, 1)

}

func (state *GameState) RelativeTile(pos Position, dir Direction, n Coordinate) (BoardTile, Position) {
	ok, relativePos := state.RelativePosition(pos, dir, n)

	if ok {
		return state.tileBoard[relativePos.row][relativePos.column], relativePos
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

func (state *GameState) FindPrefix(pos Position, dir Direction) DawgState {
	dawg := state.game.dawg
	tiles := state.tileBoard
	prefixPos := pos
	prefix := Word{}
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
		prefix = append(prefix, tile.letter)
		prefixPos = p
	}
	switch dir {
	case WEST, NORTH:
		slices.Reverse(prefix)
	}
	return dawg.FindPrefix(prefix)
}

func (state *GameState) GetNonEmptyBoardTiles(pos Position, dir Direction) Tiles {
	tiles := Tiles{}
	for {
		var ok bool
		tile := state.tileBoard[pos.row][pos.column].Tile
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

func (state *GameState) GetEmptyNonAnchorTiles(pos Position, dir Direction, maxLen Coordinate) Tiles {
	tiles := Tiles{}
	for {
		var ok bool
		tile := state.tileBoard[pos.row][pos.column].Tile
		if !state.IsTileEmpty(pos) {
			break
		}
		if state.IsAnchor(pos) {
			break
		}
		tiles = append(tiles, tile)
		if Coordinate(len(tiles)) == maxLen {
			break
		}
		ok, pos = state.AdjacentPosition(pos, dir)
		if !ok {
			break
		}
	}
	return tiles
}

func (state *GameState) IsTileEmpty(pos Position) bool {
	tile := &state.tileBoard[pos.row][pos.column]
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
	word := state.game.TilesToWord(tiles)
	return word
}

func (state *GameState) TilesToString(tiles Tiles) string {
	corpus := state.game.corpus
	var sb strings.Builder
	for _, t := range tiles {
		switch t.kind {
		case TILE_LETTER, TILE_JOKER:
			sb.WriteRune(corpus.LetterToRune(t.letter))
		}
	}
	return sb.String()
}

func (vcl *ValidCrossLetters) String(corpus Corpus) string {
	return fmt.Sprintf("{%v %s}", vcl.ok, vcl.letters.String(corpus))
}

func (bt *BoardTile) String(corpus Corpus) string {
	return fmt.Sprintf("%s validCrossLetters %s: %s %s: %s", bt.Tile.String(corpus),
		HORIZONTAL.String(), bt.validCrossLetters[HORIZONTAL].String(corpus),
		VERTICAL.String(), bt.validCrossLetters[VERTICAL].String(corpus))
}

func (state *GameState) CollectStates() GameStates {
	if state.fromState == nil {
		return GameStates{state}
	} else {
		initialStates := state.fromState.CollectStates()
		return slices.Concat(initialStates, GameStates{state})
	}
}

func (ws *WordScore) Word() Word {
	word := make(Word, len(ws.tileScores))
	for i, t := range ws.tileScores {
		word[i] = t.tile.letter
	}
	return word
}
