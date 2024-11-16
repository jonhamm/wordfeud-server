package main

import (
	"fmt"
	"io"
	"os"
	"path"
	. "wordfeud/context"
)

func WriteGameFile(game *Game, messages []string) (string, error) {
	Errorf := fmt.Errorf
	fmt := game.fmt
	fileName := GameFileName(game)
	tmpFileName := fileName + "~"
	var f *os.File
	var err error
	f, err = os.Create(tmpFileName)
	if err != nil {
		return tmpFileName, err
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name()) // clean up
		}
	}()

	switch game.options.FileFormat {
	case FILE_FORMAT_NONE:
		return "", Errorf("no file format specified for game file")
	case FILE_FORMAT_JSON:
		err = WriteGameFileJson(f, game)
	case FILE_FORMAT_TEXT:
		err = WriteGameFileText(f, game)
	case FILE_FORMAT_DEBUG:
		err = WriteGameFileDebug(f, game)
	}
	if err != nil {
		return tmpFileName, err
	}

	for _, m := range messages {
		if _, err = fmt.Fprintln(f, m); err != nil {
			return tmpFileName, err
		}
	}

	if err = f.Close(); err != nil {
		return tmpFileName, err
	}

	if err = os.Rename(tmpFileName, fileName); err != nil {
		return fileName, err
	}

	f = nil
	return fileName, nil
}

func GameFileName(game *Game) string {
	options := game.options
	filePath := path.Join(options.Directory, options.File)
	if options.Count > 1 {
		return fmt.Sprintf("%s-%d%s", filePath, game.seqno, options.FileFormat.Extension())
	} else {
		return fmt.Sprintf("%s%s", filePath, options.FileFormat.Extension())

	}
}

func WriteGameFileJson(f io.Writer, game *Game) error {
	panic("unimplemented")
}
func WriteGameFileText(f io.Writer, game *Game) error {
	fmt := game.fmt
	var err error
	if _, err = fmt.Fprintf(f, "Scrabble game %s-%d\n\n", game.options.Name, game.seqno); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(f, "Random seed: %d\nMoves: %d\n", game.RandSeed, game.nextMoveSeqNo-1); err != nil {
		return err
	}
	FprintPlayers(f, game, game.state.playerStates)
	fmt.Fprintf(f, "free tiles: (%d) %s\n", len(game.state.freeTiles), game.state.freeTiles.String(game.corpus))

	FprintBoard(f, game.board)

	states := game.CollectStates()
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

func WriteGameFileDebug(f io.Writer, game *Game) error {
	fmt := game.fmt
	var err error
	if _, err = fmt.Fprintf(f, "Scrabble game %s-%d\n\n", game.options.Name, game.seqno); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(f, "Random seed: %d\nMoves: %d\n", game.RandSeed, game.nextMoveSeqNo-1); err != nil {
		return err
	}
	FprintPlayers(f, game, game.state.playerStates)
	fmt.Fprintf(f, "free tiles: (%d) %s\n", len(game.state.freeTiles), game.state.freeTiles.String(game.corpus))

	FprintBoard(f, game.board)

	states := game.CollectStates()
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
