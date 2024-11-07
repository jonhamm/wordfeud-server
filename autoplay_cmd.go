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
	for seqno := 1; seqno <= options.count; seqno++ {
		game, err := NewGame(options, seqno, Players{BotPlayer(1), BotPlayer(2)})
		if err != nil {
			fmt.Println(result.errors(), err.Error())
			return result.result()
		}
		result.Width = int(game.width)
		result.Height = int(game.height)
		result.LetterScores = game.letterScores
		result.Board = game.board

		for n := 0; n < 1000; n++ {
			if !game.play() {
				break
			}
		}
	}
	return result.result()
}
