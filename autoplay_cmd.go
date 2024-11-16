package main

import (
	"flag"
	"fmt"
	. "wordfeud/context"
)

func autoplayCmd(options *GameOptions, args []string) *GameResult {
	result := new(GameResult)

	flag := flag.NewFlagSet("exit", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)
	for seqno := 1; seqno <= options.Count; seqno++ {
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
			if !game.Play() {
				break
			}
		}
	}
	return result.result()
}
