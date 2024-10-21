package main

import (
	"flag"
	"fmt"
)

func autoplayCmd(options *GameOptions, args []string) *GameResult {
	result := new(GameResult)

	flag := flag.NewFlagSet("exit", flag.ExitOnError)
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
	game := NewGame(options, corpus, Players{BotPlayer(1), BotPlayer(2)})
	result.Width = int(game.width)
	result.Height = int(game.height)
	result.LetterScores = game.letterScores
	result.Board = game.board

	for n := 0; n < 10; n++ {
		if !game.play() {
			break
		}
	}
	return result.result()
}
