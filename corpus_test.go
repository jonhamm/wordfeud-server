package main

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

	corpus, err := GetFileCorpus("data_test/corpus_dk_test.txt", GetLanguageAlphabet(language.Danish))
	if err != nil {
		t.Errorf("scanWordsDK() : %v", err)
		return
	}
	result_letters := make([]Word, 0)
	for _, s := range result {
		w := corpus.MakeWord(s)
		result_letters = append(result_letters, w)
	}

	words := corpus.words

	for i, w := range words {
		if !w.equal(result_letters[i]) {
			wants := runesToString(result_letters)
			got := runesToString(words)
			t.Errorf("scanWordsDK(,5) :\nwants:\n%v\ngot:\n%v\nwants[%d]: %s  got[%d]: %s", wants, got, i, corpus.wordToString(result_letters[i]), i, corpus.wordToString(words[i]))
			return
		}
	}

}

func (corpus *Corpus) letterToString(l Letter) string {
	var sb strings.Builder
	sb.WriteString("'")
	sb.WriteString(string(corpus.letterRune[l]))
	sb.WriteString("'")

	return sb.String()
}

func (corpus *Corpus) wordToString(word Word) string {
	var sb strings.Builder
	sb.WriteString("'")
	sb.WriteString(word.String(corpus))
	sb.WriteString("'")

	return sb.String()
}

func runesToString(runes []Word) string {
	var sb strings.Builder

	for i, r := range runes {
		if i > 0 {
			sb.WriteRune('\n')
		}
		sb.WriteString("    '")
		sb.WriteString(string(r))
		sb.WriteString("'")
	}
	return sb.String()
}
