package game

import (
	_ "embed"
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

func WriteGameFileHtml(game Game, gameEnded bool, messages []string) (string, error) {
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

	if err = updateGameHtml(game, dirName, gameEnded, messages); err != nil {
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

		validFilePtn := regexp.MustCompile(`^((index|move-[0-9]+)\.html)|styles\.css$`)

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

func updateGameHtml(game Game, dirName string, gameEnded bool, messages []string) error {
	var err error
	_game := game._Game()
	p := _game.fmt
	lang := game.Corpus().Language()
	states := _game.CollectStates()
	if len(states) == 0 {
		return nil
	}
	if _game.nextWriteSeqNo == 0 {
		if err = writeHtmlStyles(dirName); err != nil {
			return err
		}
	}
	if err = updateMovesHtml(states, dirName, gameEnded, p, lang); err != nil {
		return err
	}
	if err = updateIndexHtml(game._Game(), states, dirName, messages, p, lang); err != nil {
		return err
	}
	_game.nextWriteSeqNo = states[len(states)-1].move.seqno + 1
	return nil
}

func updateIndexHtml(game *_Game, states GameStates, dirName string, messages []string, p *message.Printer, lang language.Tag) error {
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

	if _, err = p.Fprintf(f, "<h2>%s %s-%d</h2>",
		Localized(lang, "Scrabble game"), game.options.Name, game.seqno); err != nil {
		return err
	}
	if _, err = p.Fprintln(f, "<h3>"); err != nil {
		return err
	}
	if _, err = p.Fprintf(f, "%s %d<br/>\n", Localized(lang, "Random number generator seed:"), game.RandSeed); err != nil {
		return err
	}
	if _, err = p.Fprintf(f, "%s %d<br/>\n", Localized(lang, "Number of moves in game:"), game.nextMoveSeqNo-1); err != nil {
		return err
	}
	for _, m := range messages {
		if _, err = p.Fprintf(f, "%s<br/>\n", m); err != nil {
			return err
		}
	}
	if _, err = p.Fprintln(f, "</h3>"); err != nil {
		return err
	}

	fmt.Fprintf(f, Localized(lang, "Remaining free tiles:")+" (%d) %s\n", len(game.state.freeTiles), game.state.freeTiles.String(game.corpus))

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
				p.Fprintf(f, "<dt><a href=\"file:move-%d.html\">%d - %s</a>", 0, 0, Localized(lang, "initial board"))
			} else {
				n := move.seqno
				p.Fprintf(f, "<dt><a href=\"file:move-%d.html\">%d - ", n, n)
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

func updateMovesHtml(states GameStates, dirName string, gameEnded bool, p *message.Printer, lang language.Tag) error {
	if len(states) == 0 {
		return nil
	}
	lastMove := states[len(states)-1].move
	lastMoveNo := uint(0)
	if lastMove != nil {
		lastMoveNo = lastMove.seqno
	}
	for _, state := range states {
		if err := updateMoveHtml(state, lastMoveNo, dirName, gameEnded, p, lang); err != nil {
			return err
		}
	}
	return nil
}

func updateMoveHtml(state *GameState, lastMove uint, dirName string, gameEnded bool, p *message.Printer, lang language.Tag) error {
	var err error
	var f *os.File
	move := state.move
	game := state.game
	corpus := game.corpus
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

	writeHtmlHeader(f, p, lang)
	p.Fprintln(f, "<body>")
	p.Fprintln(f, "<h2>")
	if seqno == 0 {
		p.Fprintln(f, Localized(lang, "Initial board"))
	} else {
		player := move.playerState.player
		word := move.state.TilesToString(move.tiles.Tiles())
		startPos := move.position
		endPos := startPos
		if game.IsValidPos(startPos) {
			_, endPos = state.RelativePosition(startPos, move.direction, Coordinate(len(word)))
		}
		p.Fprintf(f, Localized(lang, "Move number %d<br/>%s played %s %s..%s \"%s\" giving score %d")+"\n",
			move.seqno, player.name, move.direction.Orientation().Localized(lang), startPos.String(), endPos.String(), word, move.score.score)

	}
	p.Fprintln(f, "</h2>")

	p.Fprintln(f, "<p>")
	if seqno > 0 {
		p.Fprintf(f, `<a href="file:move-%d.html"><button class="navigate">%s</button></a>`, seqno-1, Localized(lang, "previous move"))
	} else {
		p.Fprintf(f, `<button class="navigate disabled">%s</button>`, Localized(lang, "previous move"))
	}
	p.Fprintln(f, "&nbsp&nbsp&nbsp\n")
	p.Fprintf(f, `<a href="file:index.html"><button class="navigate">%s</button></a>`, Localized(lang, "game overview"))
	p.Fprintln(f, "&nbsp&nbsp&nbsp\n")

	if !gameEnded || seqno < lastMove {
		p.Fprintf(f, `<a href="file:move-%d.html"><button class="navigate">%s</button></a>`, seqno+1, Localized(lang, "next move"))
	} else {
		p.Fprintf(f, `<button class="navigate disabled">%s</button>`, Localized(lang, "next move"))
	}
	p.Fprintln(f, "\n</p>")

	p.Fprintln(f, "<h3>")
	for _, ps := range state.playerStates {
		if ps.player.id != SystemPlayerId {
			p.Fprintf(f, Localized(lang, "%s has total score %d and %s")+"<br/>\n", ps.player.name, ps.score, ps.rack.Pretty(corpus))
		}
	}
	p.Fprintln(f, "</h3>")

	FprintHtmlStateBoard(f, state)

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

func writeHtmlHeader(f io.Writer, p *message.Printer, lang language.Tag) {
	p.Fprintf(f, `<!DOCTYPE html>
<html lang="%s">
<head>
    <title>HTML Other Lists</title>
    <meta charset="utf-8">
	<link rel="stylesheet" href="styles.css">
</head>
	`, lang.String())
}

func FprintHtmlStateBoard(f io.Writer, state *GameState) error {
	p := state.game.fmt
	board := state.game.board
	var err error
	corpus := state.game.corpus
	squares := board.squares
	w := board.game.dimensions.Width
	h := board.game.dimensions.Height
	tiles := state.tileBoard
	placedMinRow := int(h)
	placedMaxRow := -1
	placedMinColumn := int(w)
	placedMaxColumn := -1

	if state.move != nil {

		for _, t := range state.move.tiles {
			p := t.pos
			if t.placedInMove {
				if int(p.row) > placedMaxRow {
					placedMaxRow = int(p.row)
				}
				if int(p.row) < placedMinRow {
					placedMinRow = int(p.row)
				}
			}
			if t.placedInMove {
				if int(p.column) > placedMaxColumn {
					placedMaxColumn = int(p.column)
				}
				if int(p.column) < placedMinColumn {
					placedMinColumn = int(p.column)
				}
			}
		}

	}

	if _, err = p.Fprintln(f, `<table class="board">`); err != nil {
		return err
	}
	if _, err = p.Fprintln(f, `  <tr>`); err != nil {
		return err
	}
	if _, err = p.Fprintln(f, `    <th class="thh"/>`); err != nil {
		return err
	}
	for c := Coordinate(0); c < w; c++ {
		if _, err = p.Fprintf(f, `    <th  class="thh">%d`+"</th>\n", c); err != nil {
			return err
		}
	}
	if _, err = p.Fprintln(f, `  </tr>`); err != nil {
		return err
	}
	for r := Coordinate(0); r < h; r++ {
		if _, err = p.Fprintln(f, `  <tr>`); err != nil {
			return err
		}
		if _, err = p.Fprintf(f, `    <th  class="thv">%d`+"</th>\n", r); err != nil {
			return err
		}
		for c := Coordinate(0); c < w; c++ {
			tile := tiles[r][c]
			s := squares[r][c]
			class := ""
			k := ""
			tc := ""
			switch s {
			case DW:
				k = "dw"
			case TW:
				k = "tw"
			case DL:
				k = "dl"
			case TL:
				k = "tl"
			case CE:
				k = "ct"
			}
			switch tile.kind {
			case TILE_JOKER, TILE_LETTER:
				letterScore := board.game.GetTileScore(tile.Tile)
				tc = p.Sprintf(`<div class="tile">%s<span class="score">%d</span></div>`, tile.letter.String(corpus), letterScore)
			}

			if k != "" {
				class = p.Sprintf(` class="square %s"`, k)
			} else {
				class = p.Sprintf(` class="square"`, k)
			}
			if _, err = p.Fprintf(f, `    <td><div%s>%s</div>`+"</td>\n", class, tc); err != nil {
				return err
			}
		}
		if _, err = p.Fprintln(f, `  </tr>`); err != nil {
			return err
		}
	}

	if _, err = p.Fprintln(f, `</table>`); err != nil {
		return err
	}

	return nil
}

//go:embed styles.css
var cssStyles string

func writeHtmlStyles(dirName string) error {
	var err error
	var f *os.File

	fileName := "styles.css"
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

	if _, err = f.WriteString(cssStyles); err != nil {
		return err
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
