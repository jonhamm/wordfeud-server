package game

import (
	"io"
	. "wordfeud/localize"
)

func WriteGameFileText(f io.Writer, game Game, messages []string) error {
	_game := game._Game()
	fmt := _game.fmt
	options := _game.options
	lang := options.Language
	var err error
	if _, err = fmt.Fprintf(f, Localized(lang, "Scrabble game")+" %s-%d\n\n", _game.options.Name, _game.seqno); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(f, Localized(lang, "Random number generator seed:")+" %d\n", _game.RandSeed); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(f, Localized(lang, "Number of moves in game:")+" %d\n", _game.nextMoveSeqNo-1); err != nil {
		return err
	}

	fmt.Fprintf(f, Localized(lang, "Remaining free tiles:")+" (%d) %s\n", len(_game.state.freeTiles), _game.state.freeTiles.String(_game.corpus))

	fmt.Fprintln(f, "")

	for _, m := range messages {
		if _, err = fmt.Fprintln(f, m); err != nil {
			return err
		}
	}

	FprintBoard(f, _game.board)

	states := _game.CollectStates()
	for _, state := range states {
		fmt.Fprint(f, "\f\n")
		if state.move != nil {
			FprintMoveText(f, state.move)
		} else {
			FprintStateText(f, state)
		}
	}

	return nil
}

func FprintMoveText(f io.Writer, move *Move) {
	state := move.state
	game := state.game
	p := state.game.fmt
	corpus := state.game.corpus
	lang := corpus.Language()
	playerState := move.playerState
	player := playerState.player
	word := move.state.TilesToString(move.tiles.Tiles())
	startPos := move.position
	endPos := startPos
	if game.IsValidPos(startPos) {
		_, endPos = state.RelativePosition(startPos, move.direction, Coordinate(len(word)))
	}
	p.Fprintf(f, Localized(lang, "%s move number %d %s %s..%s \"%s\" gives score %d")+"\n\n",
		player.name, move.seqno, move.direction.Orientation().Localized(lang), startPos.String(), endPos.String(), word, move.score.score)

	for _, ps := range state.playerStates {
		if ps.player.id != SystemPlayerId {
			p.Fprintf(f, Localized(lang, "%s has total score %d and %s")+"\n", ps.player.name, ps.score, ps.rack.Pretty(corpus))

		}
	}
	p.Print("\n\n\n")
	FprintStateText(f, move.state)

}

func FprintStateText(f io.Writer, state *GameState) {
	FprintStateBoard(f, state)
}
