package main

import (
	"fmt"
	"sort"
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

func NewTestCorpusFromContent(corpus *Corpus, corpusContent []string) (*CorpusContent, error) {
	content := make([]string, len(corpusContent))
	for i, w := range corpusContent {
		content[i] = strings.ToUpper(w)
	}
	sort.Slice(content, func(i int, j int) bool {
		return strings.Compare(content[i], content[j]) < 0
	})
	data := strings.Join(content, "\n")
	return corpus.NewContent(strings.NewReader(data))
}

func testDawgContent(t *testing.T, language language.Tag, corpusContent []string) {
	corpus, err := NewCorpus(language)
	if err != nil {
		t.Errorf("testDawgContent() failed to create corpus : %v", err)
		return
	}
	content, err := NewTestCorpusFromContent(corpus, corpusContent)
	if err != nil {
		t.Errorf("testDawgContent() failed to create corpus : %v", err)
		return
	}
	testDawgCorpusContent(t, content)
}

func testDawgLanguage(t *testing.T, language language.Tag) {
	fileName, err := GetLanguageFileName(language)
	if err != nil {
		t.Errorf("testDawgContent() failed to create corpus : %v", err)
		return
	}
	testDawgLanguageFile(t, language, fileName)
}

func testDawgLanguageFile(t *testing.T, language language.Tag, fileName string) {
	corpus, err := NewCorpus(language)
	if err != nil {
		t.Errorf("testDawgContent() failed to create corpus : %v", err)
		return
	}
	content, err := corpus.GetFileContent(fileName)
	if err != nil {
		t.Errorf("testDawgLanguageFile(\"%s\") failed to create corpus : %v", fileName, err)
		return
	}
	testDawgCorpusContent(t, content)
}

func testDawgCorpusContent(t *testing.T, content *CorpusContent) {
	dawg, err := NewDawg(content)
	if err != nil {
		t.Errorf("testDawgContent() failed to create dawg : %v", err)
		return
	}
	testDawg(t, dawg, content)
}

func testDawg(t *testing.T, dawg *Dawg, content *CorpusContent) {
	fmt.Print("\n\n")
	verifyDawgCoverage(t, dawg, content)
}

func verifyDawgCoverage(t *testing.T, dawg *Dawg, content *CorpusContent) {
	verifyCorpusMatches(t, dawg, content)
	verifyMatchesInCorpus(t, dawg, content)
}

func verifyCorpusMatches(t *testing.T, dawg *Dawg, content *CorpusContent) {
	if dawg.corpus != content.corpus {
		t.Errorf("verifyCorpusMatches content.corpus != dawg.corpus")
		return
	}
	corpus := content.corpus
	for i, w := range content.words {
		if DAWG_TRACE {
			fmt.Printf("\nverifyCorpusMatches: %d \"%s\n\n", i, w.String(corpus))
		}
		s := dawg.FindPrefix(w)
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

func verifyMatchesInCorpus(t *testing.T, dawg *Dawg, content *CorpusContent) {
	if dawg.corpus != content.corpus {
		t.Errorf("verifyMatchesInCorpus content.corpus != dawg.corpus")
		return
	}
	corpus := dawg.corpus
	allWords := make(map[string]bool)
	for i, w := range content.words {
		s := w.String(corpus)
		if allWords[s] {
			t.Errorf("corpus content word #%d \"%s\" is present multiple times in content.words", i, s)
			return
		}
		allWords[s] = true
	}
	count := verifyMatchesInCorpusRecurse(t, dawg, content, allWords, dawg.initialState)

	if count != content.WordCount() {
		t.Errorf("dawg generated %v words which is not equal to corpus wordcount of %v", count, content.WordCount())
	}
}

func verifyMatchesInCorpusRecurse(t *testing.T, dawg *Dawg, content *CorpusContent, allWords map[string]bool, state DawgState) int {
	count := 0
	corpus := dawg.corpus
	node := state.LastNode()
	if node == nil {
		return count
	}
	vertex := state.LastVertex()
	if vertex != nil {
		if vertex.final {
			w := state.Word()
			s := w.String(corpus)
			_, found := content.FindWord(w)
			if !found {
				t.Errorf("dawg genrated word \"%s\" not found in corpus", s)
				return count
			}
			if !allWords[s] {
				t.Errorf("word \"%s\"  found in corpus but not in allWords", s)
				return count
			}
			delete(allWords, s)
			count++
		}
	}
	for _, v := range node.vertices {
		vs := dawg.Transition(state, v.letter)
		count += verifyMatchesInCorpusRecurse(t, dawg, content, allWords, vs)
	}
	return count
}
