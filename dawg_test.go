package main

import (
	"fmt"
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
	var dawg *Dawg
	corpus, err := MakeTestCorpusFromContent(language, corpusContent)
	if err != nil {
		t.Errorf("testDawgContent() failed to create corpus : %v", err)
		return
	}
	dawg, err = MakeDawg(corpus)
	if err != nil {
		t.Errorf("testDawgContent() failed to create dawg : %v", err)
		return
	}

	fmt.Print("\n\n")
	verifyDawgCoverage(t, dawg)
}

func verifyDawgCoverage(t *testing.T, dawg *Dawg) {
	verifyCorpusMatches(t, dawg)
	verifyMatchesInCorpus(t, dawg)
}

func verifyCorpusMatches(t *testing.T, dawg *Dawg) {
	corpus := dawg.corpus
	for i, w := range corpus.words {
		s := dawg.CommonPrefix(w)
		v := s.LastVertex()
		if v == nil {
			t.Errorf("corpus word #%v \"%s\" not matched by dawg", i, w.String(corpus))
			return
		}
		if !v.final {
			t.Errorf("corpus word #%v \"%s\" is matched by dawg but does not end in final state", i, w.String(corpus))
			return
		}
	}
}

func verifyMatchesInCorpus(t *testing.T, dawg *Dawg) {
	verifyMatchesInCorpusRecurse(t, dawg, dawg.initialState)
}

func verifyMatchesInCorpusRecurse(t *testing.T, dawg *Dawg, state State) {
	node := state.LastNode()
	if node == nil {
		return
	}
	vertex := state.LastVertex()
	if vertex != nil {
		if vertex.final {
			_, found := dawg.corpus.FindWord(state.word)
			if !found {
				t.Errorf("dawg genrated word \"%s\" not found in corpus", state.word.String(dawg.corpus))
				return
			}

		}
	}
	for _, v := range node.vertices {
		vs := dawg.Transition(state, v.letter)
		verifyMatchesInCorpusRecurse(t, dawg, vs)
	}
}
