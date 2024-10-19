package main

type Tile struct {
	letter Letter
}

type GameState struct {
	game  *Game
	tiles [][]Tile
}
