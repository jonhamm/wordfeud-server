package main

import (
	"flag"
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func gameCmd(options *GameOptions, args []string) *GameResult {
	result := new(GameResult)

	flag := flag.NewFlagSet("exit", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)
	out := options.out
	if options.debug > 0 {
		options.verbose = true
		fmt.Fprintf(out, "options: %+v\n", options)
	}

	game, err := NewGame(options, Players{BotPlayer(1), BotPlayer(2)})
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	result.Width = int(game.width)
	result.Height = int(game.height)
	result.LetterScores = game.letterScores
	result.Board = game.board
	p := message.NewPrinter(language.Danish)

	p.Fprintf(result.logger(), "Game size: width=%d height=%d squares=%d\n", game.Width(), game.Height(), game.SquareCount())
	/*
			if game.board != nil {
		   		printBoard(result.logger(), game.board)
		   	}
	*/

	state := game.state
	if state != nil {
		printState(result.logger(), state)
	}

	return result.result()
}
