package main

import (
	"reflect"
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

	words := corpus.words

	for i, w := range words {
		if !reflect.DeepEqual(result_runes[i], w) {
			wants := runesToString(result_runes)
			got := runesToString(words)
			t.Errorf("scanWordsDK(,5) :\nwants:\n%v\ngot:\n%v\nwants[%d]: %s  got[%d]: %s", wants, got, i, wordToString(result_runes[i]), i, wordToString(words[i]))
			return
		}
	}

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
