package main

import (
	"fmt"

	"golang.org/x/text/language"
)

type LanguageTile struct {
	character rune
	count     byte
	value     Score
}

type LanguageTiles []LanguageTile
type LanguageDefinition struct {
	fileName string
	pieces   LanguageTiles // string with all vowels
}

var languageDefinition = map[language.Tag]LanguageDefinition{
	language.Danish: {fileName: "corpus_dk.txt",
		pieces: LanguageTiles{
			LanguageTile{'A', 7, 1},
			LanguageTile{'B', 4, 3},
			LanguageTile{'C', 2, 8},
			LanguageTile{'D', 5, 2},
			LanguageTile{'E', 9, 1},
			LanguageTile{'F', 3, 3},
			LanguageTile{'G', 3, 3},
			LanguageTile{'H', 2, 4},
			LanguageTile{'I', 4, 3},
			LanguageTile{'J', 2, 4},
			LanguageTile{'K', 4, 3},
			LanguageTile{'L', 5, 2},
			LanguageTile{'M', 3, 4},
			LanguageTile{'N', 7, 1},
			LanguageTile{'O', 5, 2},
			LanguageTile{'P', 2, 4},
			LanguageTile{'R', 7, 1},
			LanguageTile{'S', 6, 2},
			LanguageTile{'T', 6, 2},
			LanguageTile{'U', 3, 3},
			LanguageTile{'V', 3, 4},
			LanguageTile{'X', 1, 8},
			LanguageTile{'Y', 2, 4},
			LanguageTile{'Z', 1, 9},
			LanguageTile{'Æ', 2, 4},
			LanguageTile{'Ø', 2, 4},
			LanguageTile{'Å', 2, 4},
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
