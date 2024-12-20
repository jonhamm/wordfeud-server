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

func WriteGameFileHtml(game Game, gameEnded bool, messages Messages) (string, error) {
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

		validFilePtn := regexp.MustCompile(`^((index|move-[0-9]+)\.html)|styles\.css|script\.js$`)

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

func updateGameHtml(game Game, dirName string, gameEnded bool, messages Messages) error {
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
		if err = writeHtmlScript(dirName); err != nil {
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

func updateIndexHtml(game *_Game, states GameStates, dirName string, messages Messages, p *message.Printer, lang language.Tag) error {
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
	// HEADER
	if _, err = p.Fprintln(f, `<div class="canvas">`); err != nil {
		return err
	}

	if _, err = p.Fprintf(f, `<div class="header">%s %s-%d</div>`+"\n",
		Localized(lang, "Scrabble game"), game.options.Name, game.seqno); err != nil {
		return err
	}
	p.Fprintln(f, `</div>`)

	if _, err = p.Fprintln(f, `<div class="canvas">`); err != nil {
		return err
	}
	for _, category := range AllMessageCategories {
		divId := ""
		divClass := "messages"
		showHide := false
		switch category {
		case MESSAGE_RESULT:
		case MESSAGE_DETAIL:
			divId = "DetailsShowHide"
			divClass += " hidden"
			showHide = true
		default:
			continue
		}
		if showHide {
			if _, err = p.Fprintln(f, `<div>`); err != nil {
				return err
			}
			if _, err = p.Fprintln(f, `   <button class="show-hide shown" id="DetailsShow" onClick="showDetails()">`+Localized(lang, "show details")+`</button>`); err != nil {
				return err
			}
			if _, err = p.Fprintln(f, `   <button class="show-hide hidden" id="DetailsHide" onClick="hideDetails()">`+Localized(lang, "hide details")+`</button>`); err != nil {
				return err
			}
			p.Fprintln(f, `</div>`)
		}
		if _, err = p.Fprintf(f, `<div class="%s" id="%s">`+"\n", divClass, divId); err != nil {
			return err
		}
		p.Fprintln(f, `<p/>`)
		for _, m := range messages[category] {
			if _, err = p.Fprintf(f, "%s<br/>\n", m); err != nil {
				return err
			}
		}
		p.Fprintln(f, `</div><p/>`)
	}
	p.Fprintln(f, `</div>`)
	p.Fprintln(f, `<p/>`)

	if len(states) > 0 {
		game := states[0].game
		p := game.fmt
		corpus := game.corpus
		lang := game.corpus.Language()

		p.Fprintln(f, `<dl class="move-list">`)

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

				p.Fprintf(f, Localized(lang, `%s played "%s" %s at %s scoring %d`),
					player.name, word, move.direction.Orientation().Localized(lang), startPos.String(), move.score.score)
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
	// HEADER
	p.Fprintln(f, `<div class="canvas">`)
	if seqno == 0 {
		p.Fprintln(f, `<div class="header">`+Localized(lang, "Initial board")+`</div>`)
	} else {
		player := move.playerState.player
		word := move.state.TilesToString(move.tiles.Tiles())
		startPos := move.position
		p.Fprintf(f, `<div class="header">`+Localized(lang, "Move number %d")+`</div>`, move.seqno)
		p.Fprintf(f, `<div class="move">`+Localized(lang, `%s played "%s" %s at %s giving %d points`)+`</div>`,
			player.name, word, move.direction.Orientation().Localized(lang), startPos.String(), move.score.score)
	}
	p.Fprintln(f, `</div>`)

	// Navigation buttons
	p.Fprintln(f, `<div class="canvas">`)
	p.Fprintln(f, `<table class="buttons">`)
	p.Fprintln(f, `<tr>`)
	disabled := ""
	if seqno == 0 {
		disabled = " disabled"
	}
	p.Fprintf(f, `<td><a href="file:move-%d.html"><button class="navigate"%s>&lArr;%s&lArr;</button></a></td>`+"\n",
		seqno-1, disabled, Localized(lang, "previous move"))
	p.Fprintf(f, `<td><a href="file:index.html"><button class="navigate">&uArr;%s&uArr;</button></a></td>`+"\n",
		Localized(lang, "game overview"))

	disabled = ""
	if gameEnded && seqno >= lastMove {
		disabled = " disabled"
	}
	p.Fprintf(f, `<td><a href="file:move-%d.html"><button class="navigate"%s>&rArr;%s&rArr;</button></a></td>`+"\n",
		seqno+1, disabled, Localized(lang, "next move"))

	p.Fprintln(f, `</tr>`)
	p.Fprintln(f, `</table>`)
	p.Fprintln(f, `</div>`)

	p.Fprintln(f, "\n</p>")

	if err = writeHtmlPlayerStates(f, state); err != nil {
		return err
	}

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
	<script src="script.js"></script>
</head>
	`, lang.String())
}

func writeHtmlPlayerStates(f io.Writer, state *GameState) error {
	game := state.game
	p := game.fmt
	corpus := game.corpus
	states := state.playerStates
	lang := corpus.Language()
	// Player states
	if _, err := p.Fprintln(f, `<div class="canvas">`); err != nil {
		return err
	}
	if _, err := p.Fprintln(f, `  <table class="players">`); err != nil {
		return err
	}

	for _, ps := range states {
		if ps.player.id == SystemPlayerId {
			continue
		}
		if _, err := p.Fprintln(f, `    <tr class="player">`); err != nil {
			return err
		}
		if _, err := p.Fprintf(f, `       <td><div class="name">%s</div></td>`, ps.player.name); err != nil {
			return err
		}
		if _, err := p.Fprintf(f, `       <td><div class="total-score">`+Localized(lang, "%d points")+`</div></td>`, ps.score); err != nil {
			return err
		}
		if _, err := p.Fprintln(f, `      <td>`); err != nil {
			return err
		}
		if err := writeHtmlRack(f, "          ", game, ps); err != nil {
			return err
		}
		if _, err := p.Fprintln(f, `       </td>`); err != nil {
			return err
		}
		if _, err := p.Fprintln(f, `    </tr>`); err != nil {
			return err
		}
	}

	p.Fprintln(f, `  </table>`)
	p.Fprintln(f, `</div>`)
	return nil
}

func writeHtmlRack(f io.Writer, indent string, game *_Game, ps *PlayerState) error {
	p := game.fmt
	corpus := game.corpus
	if _, err := p.Fprintf(f, `%s<table class="rack">`+"\n", indent); err != nil {
		return err
	}
	if _, err := p.Fprintf(f, `%s   <tr>`+"\n", indent); err != nil {
		return err
	}
	for _, t := range ps.rack {
		if _, err := p.Fprintf(f, `%s      <td>`, indent); err != nil {
			return err
		}
		letter := ""
		letterScore := Score(0)
		switch t.kind {
		case TILE_LETTER:
			letter = t.letter.String(corpus)
			letterScore = game.letterScores[t.letter]
		case TILE_JOKER:
			letter = "?"
		default:
		}
		if letter != "" {
			if _, err := p.Fprintf(f, `%s        <div class="square"><div class="tile">%s<span class="score">%d</span></div></div>`+"\n",
				indent, letter, letterScore); err != nil {
				return err
			}
		} else {
			if _, err := p.Fprintf(f, `%s        <div class="square"></div>`+"\n", indent); err != nil {
				return err
			}
		}

		if _, err := p.Fprintf(f, `%s      </td>`+"\n", indent); err != nil {
			return err
		}

	}

	if _, err := p.Fprintf(f, `%s   </tr>`+"\n", indent); err != nil {
		return err
	}
	if _, err := p.Fprintf(f, `%s</table>`+"\n", indent); err != nil {
		return err
	}
	return nil
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
				tileClass := "tile"
				if int(r) >= placedMinRow && int(r) <= placedMaxRow && int(c) >= placedMinColumn && int(c) <= placedMaxColumn {
					for _, t := range state.move.tiles {
						if t.pos.equal(Position{r, c}) {
							tileClass = "played"
							break
						}
					}
				}
				tc = p.Sprintf(`<div class="%s">%s<span class="score">%d</span></div>`,
					tileClass, tile.letter.String(corpus), letterScore)
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

//go:embed script.js
var script string

func writeHtmlScript(dirName string) error {
	var err error
	var f *os.File

	fileName := "script.js"
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

	if _, err = f.WriteString(script); err != nil {
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
