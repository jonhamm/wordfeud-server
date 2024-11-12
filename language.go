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

func Localized(lang language.Tag, text string) string {
	switch lang {
	case language.Danish:
		return Danish(text)
	}
	return text
}

func Danish(text string) string {
	switch text {
	case "Game completed after %d moves as %s has no more tiles in rack":
		return "Spillet afsluttet efter %d træk da %s ikke har flere bogstaver"
	case "Game completed after %d moves as there has been %d conequtive passes":
		return "Spillet afsluttet efter %d træk da der har været %d uafbrudte pas meldinger"
	case "Game is a draw between %d players: %s":
		return "Spillet er uafgjort mellem %d spillere: %s"
	case "Game is won by %s":
		return "Spillet er vundet af %s"
	case "Wrote game file after move %d \"%s\"":
		return "Skrev spil fil efter træk nummer %d \"%s\""
	case "Game file is %s":
		return "Spil filen er %s"
	case "%s scored %d with remaining rack %s":
		return "%s har scoret %d point og har disse bogstaver tilbage %s"
	}
	return text
}
