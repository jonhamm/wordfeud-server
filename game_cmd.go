package main

import (
	"flag"
	"fmt"
	. "wordfeud/context"
	. "wordfeud/game"
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
	result.Width = int(game.Dimensions().Width)
	result.Height = int(game.Dimensions().Height)
	result.LetterScores = game.LetterScores()
	result.Board = game.Board()

	game.Fmt().Fprintf(result.logger(), "Game size: width=%d height=%d squares=%d\n",
		game.Dimensions().Width, game.Dimensions().Height, game.SquareCount())
	FprintStateOfGame(result.logger(), game)

	return result.result()
}
