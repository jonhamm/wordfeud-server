package main

import (
	"flag"
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func gameCmd(options *gameOptions, args []string) *GameResult {
	result := new(GameResult)

	flag := flag.NewFlagSet("solve", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)
	out := options.out
	if options.debug > 0 {
		options.verbose = true
		fmt.Fprintf(out, "options: %+v\n", options)
	}

	corpus, err := GetLangCorpus()
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	game := NewGame(corpus, Players{BotPlayer(1), BotPlayer(2)})
	result.Width = int(game.width)
	result.Height = int(game.height)
	result.LetterScores = game.letterScores
	result.Board = game.board
	p := message.NewPrinter(language.Danish)

	p.Fprintf(result.logger(), "Game size: width=%d height=%d squares=%d\n", game.Width(), game.Height(), game.SquareCount())
	if game.board != nil {
		printBoard(result.logger(), game.board)
	}
	if game.state != nil {
		printState(result.logger(), game.state)
	}
	return result.result()
}
