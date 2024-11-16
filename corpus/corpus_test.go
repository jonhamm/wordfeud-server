package corpus

import (
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func Test_scanWordsDK(t *testing.T) {
	result := []string{
		"ab",
		"abbed",
		"abbeden",
		"abbedens",
		"abbeder",
		"abbederne",
		"abbedernes",
		"abbeders",
		"abbeds",
		"abe",
		"abede",
		"abedes",
		"abende",
		"aber",
		"abes",
		"abet",
		"bærme",
		"bærmen",
		"bærmens",
		"bærmes",
		"faldt",
		"trind",
		"trinde",
		"trindere",
		"trindest",
		"trindeste",
		"trindt",
		"æble",
		"æbler",
		"æblerne",
		"æblernes",
		"æblers",
		"æbles",
		"æblet",
		"æblets",
	}

	corpus, err := NewCorpus(language.Danish)
	if err != nil {
		t.Errorf("Test_scanWordsDK - cannot create corpus : %v", err)
		return
	}
	content, err := corpus.GetFileContent("../data_test/corpus_dk_test.txt")
	if err != nil {
		t.Errorf("Test_scanWordsDK - cannot create content : %v", err)
		return
	}

	words := content.Words()
	if len(words) != len(result) {
		t.Errorf("content has %d words but expected result has %d words", len(words), len(result))
	}

	results := make(map[string]bool)
	for _, s := range result {
		if results[s] {
			t.Errorf("expected result has \"%s\" multiple times", s)
		}
		results[strings.ToUpper(s)] = true
	}
	for _, w := range words {
		s := w.String(corpus)
		if !results[s] {
			t.Errorf("result has \"%s\" which is not an expected result", s)
			return
		}
		results[s] = false
	}
}
