package game

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
)

func WriteGameFileHtml(game Game, messages []string) (string, error) {
	Errorf := fmt.Errorf
	_game := game._Game()
	state := _game.state
	dirName := GameFileName(game)
	var err error
	switch dirName {
	case "", ".", "..", "/":
		return "", Errorf("invalid dir name for HTML directory: \"%s\"", dirName)
	}
	if state.move.seqno < _game.nextWriteSeqNo {
		// written .html files are up to date
		return dirName, nil
	}

	if _game.nextWriteSeqNo == 0 {
		if err = rmHtmlDir(dirName); err != nil {
			return dirName, err
		}

		err = os.MkdirAll(dirName, 0777)
		if err != nil {
			return "", err
		}
	}

	if err = updateGameHtml(game, dirName); err != nil {
		return "", err
	}

	return dirName, nil
}

func rmHtmlDir(dirName string) error {
	Errorf := fmt.Errorf
	info, err := os.Stat(dirName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		files, err := os.ReadDir(dirName)
		if err != nil {
			return err
		}

		validFilePtn := regexp.MustCompile(`^(index|move-[0-9]+)\.html$`)

		for _, file := range files {
			if !validFilePtn.Match([]byte(file.Name())) {
				return Errorf("existing directory \"%s\" contains files other than game .html files: \"%s\"", file.Name())
			}
		}

		// dirName directory exists but contains only index.html and move-nnn.html files
		// it is deemed safe to remove this directory
		if err = os.RemoveAll(dirName); err != nil {
			return err
		}
	} else {
		// dirName is a file
		// it is deemed safe to remove this file
		if err = os.Remove(dirName); err != nil {
			return err
		}
	}
	return nil
}

func updateGameHtml(game Game, dirName string) error {
	var err error
	_game := game._Game()
	states := _game.CollectStates()
	if len(states) == 0 {
		return nil
	}
	if err = updateMovesHtml(states, dirName); err != nil {
		return err
	}
	if err = updateIndexHtml(states, dirName); err != nil {
		return err
	}
	_game.nextWriteSeqNo = states[len(states)-1].move.seqno + 1
	return nil
}

func updateIndexHtml(states GameStates, dirName string) error {
	var err error
	var f *os.File

	fileName := "index.html"
	tmpFileName := fileName + "~"
	filePath := path.Join(dirName, fileName)
	tmpFilePath := path.Join(dirName, tmpFileName)
	if f, err = os.Create(tmpFilePath); err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name()) // clean up
		}
	}()

	for _, state := range states {
		n := uint(0)
		if state.move != nil {
			n = state.move.seqno
		}
		fmt.Fprintf(f, "<a href=\"file:move-%d.html\">Move number %d</a>\n", n, n)
	}

	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}

	f = nil
	return nil
}

func updateMovesHtml(states GameStates, dirName string) error {
	for _, state := range states {
		if err := updateMoveHtml(state, dirName); err != nil {
			return err
		}
	}
	return nil
}

func updateMoveHtml(state *GameState, dirName string) error {
	var err error
	var f *os.File
	move := state.move
	seqno := uint(0)
	if move != nil {
		seqno = move.seqno
	}
	if seqno < state.game.nextWriteSeqNo {
		return nil
	}
	fileName := fmt.Sprintf("move-%d.html", seqno)
	tmpFileName := fileName + "~"
	filePath := path.Join(dirName, fileName)
	tmpFilePath := path.Join(dirName, tmpFileName)
	if f, err = os.Create(tmpFilePath); err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name()) // clean up
		}
	}()

	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}

	f = nil
	return nil
}
