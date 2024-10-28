package main

import (
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func Test_DawgBasic(t *testing.T) {
	content := []string{
		"abe",
		"abede",
		"abedes",
		"abefest",
		"abefesten",
		"abefestens",
		"abefester",
		"abefesterne",
		"abekat",
		"bil",
		"biler",
		"bilers",
	}
	testDawgContent(t, language.Danish, content)
}

func MakeTestCorpusFromContent(language language.Tag, content []string) (*Corpus, error) {
	data := strings.Join(content, "\n")
	return NewCorpus(strings.NewReader(data), GetLanguageAlphabet(language))
}

func testDawgContent(t *testing.T, language language.Tag, corpusContent []string) {
	corpus, err := MakeTestCorpusFromContent(language, corpusContent)
	if err != nil {
		t.Errorf("testDawgContent() failed to create corpus : %v", err)
		return
	}
	_, err = MakeDawg(corpus)
	if err != nil {
		t.Errorf("testDawgContent() failed to create dawg : %v", err)
		return
	}
}
