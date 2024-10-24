package main

import (
	"fmt"

	"golang.org/x/text/language"
)

type LanguageTile struct {
	character rune
	value     Score
	count     byte
}

type LanguageTiles []LanguageTile
type LanguageDefinition struct {
	fileName string
	pieces   LanguageTiles // string with all vowels
}

var languageDefinition = map[language.Tag]LanguageDefinition{
	language.Danish: {fileName: "corpus_dk.txt",
		pieces: LanguageTiles{
			LanguageTile{'a', 7, 1},
			LanguageTile{'b', 4, 3},
			LanguageTile{'c', 2, 8},
			LanguageTile{'d', 5, 2},
			LanguageTile{'e', 9, 1},
			LanguageTile{'f', 3, 3},
			LanguageTile{'g', 3, 3},
			LanguageTile{'h', 2, 4},
			LanguageTile{'i', 4, 3},
			LanguageTile{'j', 2, 4},
			LanguageTile{'k', 4, 3},
			LanguageTile{'l', 5, 2},
			LanguageTile{'m', 3, 4},
			LanguageTile{'n', 7, 1},
			LanguageTile{'o', 5, 2},
			LanguageTile{'p', 2, 4},
			LanguageTile{'r', 7, 1},
			LanguageTile{'s', 6, 2},
			LanguageTile{'t', 6, 2},
			LanguageTile{'u', 3, 3},
			LanguageTile{'v', 3, 4},
			LanguageTile{'x', 1, 8},
			LanguageTile{'y', 2, 4},
			LanguageTile{'z', 1, 9},
			LanguageTile{'æ', 2, 4},
			LanguageTile{'ø', 2, 4},
			LanguageTile{'å', 2, 4},
		}},

	//language.English: {fileName: "corpus_en.txt", validCharacters: "a-z", vowels: "aeiouy"},
}

func GetLanguageCorpus(language language.Tag) (*Corpus, error) {
	definition, ok := languageDefinition[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language %s", language.String())
	}
	return GetFileCorpus(fmt.Sprintf("data/%s", definition.fileName), GetLanguageAlphabet(language))
}

func GetLanguageFileName(language language.Tag) string {
	definition, ok := languageDefinition[language]
	if ok {
		return fmt.Sprintf("data/%s", definition.fileName)
	}
	return ""
}

func GetLanguageTiles(language language.Tag) LanguageTiles {
	definition, ok := languageDefinition[language]
	if ok {
		return definition.pieces
	}
	return LanguageTiles{}
}

func GetLanguageAlphabet(language language.Tag) Alphabet {
	definition, ok := languageDefinition[language]
	if ok {
		letters := make(Alphabet, len(definition.pieces))
		for i, p := range definition.pieces {
			letters[i] = p.character
		}
		return letters
	}
	return Alphabet{}
}
