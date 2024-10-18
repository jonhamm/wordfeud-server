package main

import (
	"fmt"

	"golang.org/x/text/language"
)

type LanguagePiece struct {
	character   rune
	value       byte
	initalCount byte
}

type LanguagePieces []LanguagePiece
type LanguageDefinition struct {
	fileName string
	pieces   LanguagePieces // string with all vowels
}

var languageDefinition = map[language.Tag]LanguageDefinition{
	language.Danish: {fileName: "corpus_dk.txt",
		pieces: LanguagePieces{
			LanguagePiece{'a', 7, 1},
			LanguagePiece{'b', 4, 3},
			LanguagePiece{'c', 2, 8},
			LanguagePiece{'d', 5, 2},
			LanguagePiece{'e', 9, 1},
			LanguagePiece{'f', 3, 3},
			LanguagePiece{'g', 3, 3},
			LanguagePiece{'h', 2, 4},
			LanguagePiece{'i', 4, 3},
			LanguagePiece{'j', 2, 4},
			LanguagePiece{'k', 4, 3},
			LanguagePiece{'l', 5, 2},
			LanguagePiece{'m', 3, 4},
			LanguagePiece{'n', 7, 1},
			LanguagePiece{'o', 5, 2},
			LanguagePiece{'p', 2, 4},
			LanguagePiece{'r', 7, 1},
			LanguagePiece{'s', 6, 2},
			LanguagePiece{'t', 6, 2},
			LanguagePiece{'u', 3, 3},
			LanguagePiece{'v', 3, 4},
			LanguagePiece{'x', 1, 8},
			LanguagePiece{'y', 2, 4},
			LanguagePiece{'z', 1, 9},
			LanguagePiece{'æ', 2, 4},
			LanguagePiece{'ø', 2, 4},
			LanguagePiece{'å', 2, 4},
		}},

	//language.English: {fileName: "corpus_en.txt", validCharacters: "a-z", vowels: "aeiouy"},
}

func GetLangCorpus() (*Corpus, error) {
	language := language.Danish
	definition, ok := languageDefinition[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language %s", language.String())
	}
	return GetFileCorpus(fmt.Sprintf("data/%s", definition.fileName), definition.pieces)
}

func GetLanguagePieces(language language.Tag) LanguagePieces {
	definition, ok := languageDefinition[language]
	if ok {
		return definition.pieces
	}
	return LanguagePieces{}
}
