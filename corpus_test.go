package main

import (
	"slices"
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

	result_runes := make([][]rune, 0)
	for _, s := range result {
		w := MakeWord(s)
		result_runes = append(result_runes, w)
	}
	corpus, err := GetFileCorpus("data_test/corpus_dk_test.txt", GetLanguagePieces(language.Danish))
	if err != nil {
		t.Errorf("scanWordsDK() : %v", err)
		return
	}
	testCorpusStatistics(t, corpus)

	words := corpus.words

	for i, w := range words {
		if !equalWord(result_runes[i], w) {
			wants := runesToString(result_runes)
			got := runesToString(words)
			t.Errorf("scanWordsDK(,5) :\nwants:\n%v\ngot:\n%v\nwants[%d]: %s  got[%d]: %s", wants, got, i, wordToString(result_runes[i]), i, wordToString(words[i]))
			return
		}
	}

}

func Test_CorpusStatistics(t *testing.T) {
	corpus, err := GetLangCorpus()
	if err != nil {
		t.Errorf("CorpusStatistics() : %v", err)
		return
	}
	testCorpusStatistics(t, corpus)
}

func testCorpusStatistics(t *testing.T, corpus *Corpus) {
	maxWordLength := corpus.MaxWordLength()
	wordOccurence := make([]int, corpus.wordCount)
	wordsByLengthCount := 0
	for length := 1; length <= maxWordLength; length++ {
		index := corpus.GetWordLengthIndex(length)
		wordsByLengthCount += len(index)
		for _, x := range index {
			if x < 0 || x >= corpus.wordCount {
				t.Errorf("CorpusStatistics() wordLengthIndex[%d] has index entry %d which is out of rage %d..%d", length, x, 0, corpus.wordCount)
				return
			}
			if wordOccurence[x] != 0 {
				t.Errorf("CorpusStatistics() wordLengthIndex[%d] has index entry %d which is also present in  wordLengthIndex[%d]", length, x, wordOccurence[x])
				return
			}

			wordOccurence[x] = length
		}
	}
	for i, w := range corpus.words {
		for j, r := range w {
			index := corpus.GetPositionIndex(r, j)
			if !slices.Contains(index, i) {
				t.Errorf("CorpusStatistics() positionIndex[%s,%d] has no entry %d for word[%d]: %s", runeToString(r), j, i, i, wordToString(w))
				return
			}
		}
	}

}

func runeToString(r rune) string {
	var sb strings.Builder
	sb.WriteString("'")
	sb.WriteString(string(r))
	sb.WriteString("'")

	return sb.String()
}
func wordToString(runes []rune) string {
	var sb strings.Builder
	sb.WriteString("'")
	sb.WriteString(string(runes))
	sb.WriteString("'")

	return sb.String()
}

func runesToString(runes [][]rune) string {
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
