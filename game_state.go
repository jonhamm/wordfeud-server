package main

type Tile struct {
	joker  bool
	letter Letter
}

type GameState struct {
	game  *Game
	tiles [][]Tile
}

type Move struct {
	player     *Player
	board      *Board
	row        byte
	column     byte
	horizontal bool
	word       Word
}
type MoveResult struct {
	fromBoard *Board
	move      *Move
	score     int
	toBoard   *Board
}
