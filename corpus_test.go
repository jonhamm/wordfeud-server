package main

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func NewTestCorpus(wordSize int, content ...string) (*Corpus, error) {
	data := strings.Join(content, "\n")
	return NewCorpus(language.Danish, wordSize, strings.NewReader(data))
}

func Test_scanWordsDK(t *testing.T) {
	result_5 := [][]rune{
		[]rune("abbed"),
		[]rune("abede"),
		[]rune("bærme"),
		[]rune("faldt"),
		[]rune("trind"),
		[]rune("æbler"),
		[]rune("æbles"),
		[]rune("æblet")}
	corpus, err := GetFileCorpus("data_test/corpus_dk_test.txt", language.Danish, 5)
	if err != nil {
		t.Errorf("scanWordsDK() : %v", err)
		return
	}

	words := corpus.words

	if !reflect.DeepEqual(result_5, words) {
		wants := runesToString(result_5)
		got := runesToString(words)
		t.Errorf("scanWordsDK(,5) :\nwants:\n%v\ngot:\n%v", wants, got)
	}

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
