package main

type Square struct {
	row     byte
	column  byte
	kind    byte
	content rune
}

type Board struct {
	game    *Game
	squares [][]Square
}

func NewBoard(game *Game) *Board {
	board := Board{
		game:    game,
		squares: [][]Square{},
	}
	board.squares = make([][]Square, game.height)
	for i := range board.squares {
		board.squares[i] = make([]Square, game.width)
	}
	var r byte
	var c byte
	for r = 0; r < game.height; r++ {
		for c = 0; c < game.width; c++ {
			square := &board.squares[r][c]
			square.row = r
			square.column = c
			square.kind = 0
			square.content = EMPTY
		}
	}
	return &board
}
