package main

import "slices"

type Move struct {
	state       *GameState
	playerState *PlayerState
	position    Position
	direction   Direction
	word        Word
	score       Score
}

type MoveContext struct {
	state   *GameState
	rack    Rack
	anchors Positions
}

func MakeMoveContext(state *GameState, rack Rack, anchors Positions) *MoveContext {
	return &MoveContext{
		state:   state,
		rack:    slices.Clone(rack),
		anchors: slices.Clone(anchors),
	}
}

func MakeMove(context *MoveContext, postion Position, direction Direction, word Word, score Score) *Move {
	//return &Move{context.state, context.playerState, postion, direction, word, score}
	return nil
}

func (context *MoveContext) NextMove() *Move {
	return nil
}
