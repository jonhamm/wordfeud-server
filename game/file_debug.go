package game

import "io"

func WriteGameFileDebug(f io.Writer, game Game, messages []string) error {
	_game := game._Game()
	fmt := _game.fmt
	var err error
	if _, err = fmt.Fprintf(f, "Scrabble game %s-%d\n\n", _game.options.Name, _game.seqno); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(f, "Random seed: %d\nMoves: %d\n", _game.RandSeed, _game.nextMoveSeqNo-1); err != nil {
		return err
	}

	fmt.Fprintln(f, "")

	for _, m := range messages {
		if _, err = fmt.Fprintln(f, m); err != nil {
			return err
		}
	}

	FprintPlayers(f, game, _game.state.playerStates)
	fmt.Fprintf(f, "free tiles: (%d) %s\n", len(_game.state.freeTiles), _game.state.freeTiles.String(_game.corpus))

	FprintBoard(f, _game.board)

	states := _game.CollectStates()
	for _, state := range states {
		fmt.Fprint(f, "\f\n")
		if state.move != nil {
			FprintMove(f, state.move)
		} else {
			FprintState(f, state)
		}
	}

	return nil
}
