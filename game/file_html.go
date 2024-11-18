package game

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	. "wordfeud/localize"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func WriteGameFileHtml(game Game, messages []string) (string, error) {
	_game := game._Game()
	Errorf := fmt.Errorf
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
	p := _game.fmt
	lang := game.Corpus().Language()
	states := _game.CollectStates()
	if len(states) == 0 {
		return nil
	}
	if err = updateMovesHtml(states, dirName, p, lang); err != nil {
		return err
	}
	if err = updateIndexHtml(states, dirName, p, lang); err != nil {
		return err
	}
	_game.nextWriteSeqNo = states[len(states)-1].move.seqno + 1
	return nil
}

func updateIndexHtml(states GameStates, dirName string, p *message.Printer, lang language.Tag) error {
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

	writeHtmlHeader(f, p, lang)
	p.Fprintln(f, "<body>")
	if len(states) > 0 {
		game := states[0].game
		p := game.fmt
		corpus := game.corpus
		lang := game.corpus.Language()

		p.Fprintln(f, "<dl>")

		for _, state := range states {
			move := state.move
			p.Fprintln(f, "<dt>")
			if move == nil {
				p.Fprintf(f, "<dt><a href=\"file:move-%d.html\">%d: %s</a>", 0, 0, Localized(lang, "initial board"))
			} else {
				n := move.seqno
				p.Fprintf(f, "<dt><a href=\"file:move-%d.html\">%d: ", n, n)
				player := move.playerState.player
				word := move.state.TilesToString(move.tiles.Tiles())
				startPos := move.position
				endPos := startPos
				if game.IsValidPos(startPos) {
					_, endPos = state.RelativePosition(startPos, move.direction, Coordinate(len(word)))
				}
				p.Fprintf(f, Localized(lang, "%s played %s %s..%s \"%s\" giving score %d"),
					player.name, move.direction.Orientation().Localized(lang), startPos.String(), endPos.String(), word, move.score.score)
				p.Fprintln(f, "</a>")
			}
			p.Fprintln(f, "</dt>")
			p.Fprintln(f, "<dd>")
			for _, ps := range state.playerStates {
				if ps.player.id != SystemPlayerId {
					p.Fprintf(f, Localized(lang, "%s has total score %d and %s")+"<br/>\n", ps.player.name, ps.score, ps.rack.Pretty(corpus))

				}
			}
			p.Fprintln(f, "</dd>")

		}
		p.Fprintln(f, "</dl>")
	}
	p.Fprintln(f, "</body>")

	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}

	f = nil
	return nil
}

func updateMovesHtml(states GameStates, dirName string, p *message.Printer, lang language.Tag) error {
	for _, state := range states {
		if err := updateMoveHtml(state, dirName, p, lang); err != nil {
			return err
		}
	}
	return nil
}

func updateMoveHtml(state *GameState, dirName string, p *message.Printer, lang language.Tag) error {
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
	fileName := p.Sprintf("move-%d.html", seqno)
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

func writeHtmlHeader(f io.Writer, p *message.Printer, lang language.Tag) {
	p.Fprintf(f, `<!DOCTYPE html>
<html lang="%s">
<head>
    <title>HTML Other Lists</title>
    <meta charset="utf-8">
</head>
	`, lang.String())
}
