package game

import (
	"fmt"
	"os"
	"path"
	. "wordfeud/context"
)

func WriteGameFile(game Game, messages []string) (string, error) {
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
		err = WriteGameFileHtml(f, game, messages)
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
	filePath := path.Join(options.Directory, options.File)
	if options.Count > 1 {
		return fmt.Sprintf("%s-%d%s", filePath, _game.seqno, options.FileFormat.Extension())
	} else {
		return fmt.Sprintf("%s%s", filePath, options.FileFormat.Extension())

	}
}
