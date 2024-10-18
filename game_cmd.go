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
	game := NewGame(corpus)
	result.Width = int(game.width)
	result.Height = int(game.height)
	result.Pieces = game.pieces
	if len(game.moves) > 0 {
		result.StartBoard = game.moves[0].toBoard
	}

	p := message.NewPrinter(language.Danish)

	p.Fprintf(result.logger(), "Game size: width=%d height=%d squares=%d\n", game.Width(), game.Height(), game.SquareCount())
	if result.StartBoard != nil {
		printBoard(result.logger(), result.StartBoard)
	}
	return result.result()
}
