package main

import (
	"flag"
	"fmt"
)

func gameCmd(options *GameOptions, args []string) *GameResult {
	result := new(GameResult)

	flag := flag.NewFlagSet("exit", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)

	game, err := NewGame(options, 1, Players{BotPlayer(1), BotPlayer(2)})
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	result.Width = int(game.width)
	result.Height = int(game.height)
	result.LetterScores = game.letterScores
	result.Board = game.board

	game.fmt.Fprintf(result.logger(), "Game size: width=%d height=%d squares=%d\n", game.Width(), game.Height(), game.SquareCount())
	if game.state != nil {
		FprintState(result.logger(), game.state)
	}

	return result.result()
}
