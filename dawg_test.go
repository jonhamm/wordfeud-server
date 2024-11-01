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

func Test_DawgPartialDK(t *testing.T) {
	testDawgLanguageFile(t, language.Danish, "data_test/dk_partial.txt")
}

func Test_DawgCompleteDK(t *testing.T) {
	testDawgLanguage(t, language.Danish)
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
	testDawgCorpus(t, corpus)
}

func testDawgLanguage(t *testing.T, language language.Tag) {
	testDawgLanguageFile(t, language, GetLanguageFileName(language))
}

func testDawgLanguageFile(t *testing.T, language language.Tag, fileName string) {
	corpus, err := GetFileCorpus(fileName, GetLanguageAlphabet((language)))
	if err != nil {
		t.Errorf("testDawgLanguageFile(\"%s\") failed to create corpus : %v", fileName, err)
		return
	}
	testDawgCorpus(t, corpus)
}

func testDawgCorpus(t *testing.T, corpus *Corpus) {
	dawg, err := MakeDawg(corpus)
	if err != nil {
		t.Errorf("testDawgContent() failed to create dawg : %v", err)
		return
	}
	testDawg(t, dawg)
}

func testDawg(t *testing.T, dawg *Dawg) {
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
		if DAWG_TRACE {
			fmt.Printf("\nverifyCorpusMatches: %d \"%s\n\n", i, w.String(corpus))
		}
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
	corpus := dawg.corpus
	count := verifyMatchesInCorpusRecurse(t, dawg, dawg.initialState)
	if count != corpus.wordCount {
		t.Errorf("dawg generated %v words which is not equal to corpus wordcount of %v", count, corpus.wordCount)
	}
}

func verifyMatchesInCorpusRecurse(t *testing.T, dawg *Dawg, state State) int {
	count := 0
	corpus := dawg.corpus
	node := state.LastNode()
	if node == nil {
		return count
	}
	vertex := state.LastVertex()
	if vertex != nil {
		if vertex.final {
			_, found := corpus.FindWord(state.word)
			if !found {
				t.Errorf("dawg genrated word \"%s\" not found in corpus", state.word.String(corpus))
				return count
			}
			count++
		}
	}
	for _, v := range node.vertices {
		vs := dawg.Transition(state, v.letter)
		count += verifyMatchesInCorpusRecurse(t, dawg, vs)
	}
	return count
}
