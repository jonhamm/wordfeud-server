package main

import (
	"fmt"

	"golang.org/x/text/language"
)

type LanguageDefinition struct {
	fileName        string
	validCharacters string // regular expression for valid characters
	vowels          string // string with all vowels
}

var languageMatcher = language.NewMatcher([]language.Tag{
	language.Danish,
	language.English,
})

var languageDefinition = map[language.Tag]LanguageDefinition{
	language.Danish:  {fileName: "corpus_dk.txt", validCharacters: "a-zæøå", vowels: "aeiouyæøå"},
	language.English: {fileName: "corpus_en.txt", validCharacters: "a-z", vowels: "aeiouy"},
}

func GetLangCorpus(language language.Tag, wordSize int) (*Corpus, error) {
	definition, ok := languageDefinition[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language %s", language.String())
	}
	return GetFileCorpus(fmt.Sprintf("data/%s", definition.fileName), language, wordSize)
}

func GetValidLanguageCharacters(language language.Tag) string {
	definition, ok := languageDefinition[language]
	if ok {
		return definition.validCharacters
	}
	return "\\p{Ll}"
}

func GetLanguageVowels(language language.Tag) string {
	definition, ok := languageDefinition[language]
	if ok {
		return definition.vowels
	}
	return ""
}
