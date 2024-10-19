package main

import (
	"strings"
)

type ActionResult struct {
	Err     []string        `json:"err"`
	Log     []string        `json:"log"`
	_logger strings.Builder `json:"-"`
	_errors strings.Builder `json:"-"`
}

type ActionResultLogger struct {
	actionResult *ActionResult
}

func (w ActionResultLogger) Write(data []byte) (int, error) {
	return w.actionResult._logger.Write(data)
}

func (w ActionResultLogger) WriteString(str string) (int, error) {
	return w.actionResult._logger.WriteString(str)
}

type ActionResultErrorLogger struct {
	actionResult *ActionResult
}

func (w ActionResultErrorLogger) Write(data []byte) (int, error) {
	ln, lerr := w.actionResult._logger.Write(data)
	en, eerr := w.actionResult._errors.Write(data)
	if lerr != nil {
		return ln, lerr
	}

	if eerr != nil {
		return en, eerr
	}
	return en, nil
}

func (w ActionResultErrorLogger) WriteString(str string) (int, error) {
	ln, lerr := w.actionResult._logger.WriteString(str)
	en, eerr := w.actionResult._errors.WriteString(str)
	if lerr != nil {
		return ln, lerr
	}

	if eerr != nil {
		return en, eerr
	}
	return en, nil
}

type CorpusResult struct {
	ActionResult
	Words           []string                               `json:"words"`
	WordCount       int                                    `json:"wordCount"`     // total number of words - i.e. len(Words)
	MaxWordLength   int                                    `json:"maxWordLength"` // the longest word in Words
	WordLengthIndex [] /*word length*/ [] /*word ids*/ int `json:"wordLengthIndex"`
}

type GameResult struct {
	ActionResult
	Width        int          `json:"width"`
	Height       int          `json:"height"`
	LetterScores LetterScores `json:"pieceValues"`
	Tiles        []rune       `json:"tiles"`
	Board        *Board       `json:"board"`
}

func (a *ActionResult) logger() ActionResultLogger {
	return ActionResultLogger{a}
}
func (a *ActionResult) errors() ActionResultErrorLogger {
	return ActionResultErrorLogger{a}
}

func (a *ActionResult) setResult() {
	a.Err = strings.Split(a._errors.String(), "\n")
	a.Log = strings.Split(a._logger.String(), "\n")
}

func (r *CorpusResult) result() *CorpusResult {
	r.setResult()
	return r
}
func (r *GameResult) result() *GameResult {
	r.setResult()
	return r
}
