package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type FileFormat byte

const (
	FILE_FORMAT_NONE = FileFormat(0)
	FILE_FORMAT_TEXT = FileFormat(1)
	FILE_FORMAT_JSON = FileFormat(2)
)

func (format FileFormat) Extension() string {
	switch format {
	case FILE_FORMAT_NONE:
		return ""
	case FILE_FORMAT_TEXT:
		return ".txt"
	case FILE_FORMAT_JSON:
		return ".json"
	}
	panic(fmt.Sprintf("illegal FileFormat %d (FileFormat.Extension)", format))
}

func ParseFileFormat(formatSpec string) FileFormat {
	switch strings.ToLower(formatSpec) {
	case "txt", "text":
		return FILE_FORMAT_TEXT
	case "json", "jsn":
		return FILE_FORMAT_JSON
	}
	return FILE_FORMAT_NONE
}

func (format FileFormat) String() string {

	switch format {
	case FILE_FORMAT_NONE:
		return "none"
	case FILE_FORMAT_TEXT:
		return "text"
	case FILE_FORMAT_JSON:
		return "json"
	}
	panic(fmt.Sprintf("illegal FileFormat %d (FileFormat.String)", format))

}

func WriteGameFile(game *Game) (string, error) {
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
			os.Remove(f.Name()) // clean up
		}
	}()

	switch game.options.fileFormat {
	case FILE_FORMAT_NONE:
		return "", fmt.Errorf("no file format specified for game file")
	case FILE_FORMAT_JSON:
		err = WriteGameFileJson(f, game)
	case FILE_FORMAT_TEXT:
		err = WriteGameFileText(f, game)
	}
	if err != nil {
		return tmpFileName, err
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
	filePath := path.Join(options.directory, options.file)
	if options.count > 1 {
		return fmt.Sprintf("%s-%d%s", filePath, game.seqno, options.fileFormat.Extension())
	} else {
		return fmt.Sprintf("%s%s", filePath, options.fileFormat.Extension())

	}
}

func WriteGameFileJson(f io.Writer, game *Game) error {
	panic("unimplemented")
}

func WriteGameFileText(f io.Writer, game *Game) error {
	fmt := game.fmt
	var err error
	if _, err = fmt.Fprintf(f, "Scrabble game %s-%d\n\n", game.options.name, game.seqno); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(f, "Random seed: %d\nMoves: %d\n", game.randSeed, game.nextMoveSeqNo-1); err != nil {
		return err
	}
	fprintPlayers(f, game, game.state.playerStates)
	fmt.Fprintf(f, "free tiles: (%d) %s\n", len(game.state.freeTiles), game.state.freeTiles.String(game.corpus))

	fprintBoard(f, game.board)

	states := game.CollectStates()
	for _, state := range states {
		fmt.Fprint(f, "\f\n")
		if state.move != nil {
			fprintMove(f, state.move)
		} else {
			fprintState(f, state)
		}
	}

	return nil
}
