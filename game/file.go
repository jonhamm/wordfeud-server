package game

import (
	"fmt"
	"os"
	. "wordfeud/context"
)

func WriteGameFile(game Game, gameEnded bool, messages Messages) (string, error) {
	Errorf := fmt.Errorf
	options := game.Options()
	var err error
	var fileName = ""
	if err = os.MkdirAll(options.Directory, 0777); err != nil {
		return "", err
	}
	switch game.Options().FileFormat {
	case FILE_FORMAT_NONE:
		err = Errorf("no file format specified for game file")
	case FILE_FORMAT_JSON, FILE_FORMAT_TEXT, FILE_FORMAT_DEBUG:
		fileName, err = WriteFile(game, messages)
	case FILE_FORMAT_HTML:
		fileName, err = WriteGameFileHtml(game, gameEnded, messages)
	default:
		panic(fmt.Sprintf("unknown file format: %d", game.Options().FileFormat))
	}
	return fileName, err
}

func WriteFile(game Game, messages Messages) (string, error) {
	Errorf := fmt.Errorf
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

	switch game.Options().FileFormat {
	case FILE_FORMAT_NONE:
		return "", Errorf("no file format specified for game file")
	case FILE_FORMAT_JSON:
		err = WriteGameFileJson(f, game, messages)
	case FILE_FORMAT_TEXT:
		err = WriteGameFileText(f, game, messages)
	case FILE_FORMAT_HTML:
		panic("invalid file format HTML specified")
	case FILE_FORMAT_DEBUG:
		err = WriteGameFileDebug(f, game, messages)
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

func GameFileName(game Game) string {
	_game := game._Game()
	options := _game.options
	if options.Count > 1 {
		return fmt.Sprintf("%s-%d%s", options.File, _game.seqno, options.FileFormat.Extension())
	} else {
		return fmt.Sprintf("%s%s", options.File, options.FileFormat.Extension())

	}
}
