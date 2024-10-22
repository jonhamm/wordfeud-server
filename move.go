package main

type Move struct {
	state       *GameState
	playerState *PlayerState
	position    Position
	direction   Direction
	word        Word
	score       Score
}

type MoveContext struct {
	state       *GameState
	playerState *PlayerState
}

func MakeMoveContext(state *GameState,
	playerState *PlayerState) *MoveContext {
	return &MoveContext{state, playerState}
}

func MakeMove(context *MoveContext, postion Position, direction Direction, word Word, score Score) *Move {
	return &Move{context.state, context.playerState, postion, direction, word, score}
}

func (context *MoveContext) NextMove() *Move {
	return nil
}
