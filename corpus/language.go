package corpus

import (
	"fmt"
	"sort"
	"unicode"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type languageTile struct {
	character rune
	count     byte
	value     byte
}

type LanguageTiles []languageTile
type languageDefinition struct {
	language    language.Tag
	initialized bool
	collator    *collate.Collator
	alphabet    Alphabet
	fileName    string
	pieces      LanguageTiles // string with all vowels
}

var languageDefinitions = map[language.Tag]*languageDefinition{
	language.Danish: {
		language:    language.Danish,
		initialized: false,
		collator:    nil,
		alphabet:    Alphabet{},
		fileName:    "corpus_dk.txt",
		pieces: LanguageTiles{
			languageTile{'A', 7, 1},
			languageTile{'B', 4, 3},
			languageTile{'C', 2, 8},
			languageTile{'D', 5, 2},
			languageTile{'E', 9, 1},
			languageTile{'F', 3, 3},
			languageTile{'G', 3, 3},
			languageTile{'H', 2, 4},
			languageTile{'I', 4, 3},
			languageTile{'J', 2, 4},
			languageTile{'K', 4, 3},
			languageTile{'L', 5, 2},
			languageTile{'M', 3, 4},
			languageTile{'N', 7, 1},
			languageTile{'O', 5, 2},
			languageTile{'P', 2, 4},
			languageTile{'R', 7, 1},
			languageTile{'S', 6, 2},
			languageTile{'T', 6, 2},
			languageTile{'U', 3, 3},
			languageTile{'V', 3, 4},
			languageTile{'X', 1, 8},
			languageTile{'Y', 2, 4},
			languageTile{'Z', 1, 9},
			languageTile{'Æ', 2, 4},
			languageTile{'Ø', 2, 4},
			languageTile{'Å', 2, 4},
		}},

	//language.English: {fileName: "corpus_en.txt", validCharacters: "a-z", vowels: "aeiouy"},
}

func (def *languageDefinition) init() {
	if def.initialized {
		return
	}
	def.collator = collate.New(def.language)
	def.fileName = fmt.Sprintf("data/%s", def.fileName)
	characters := make([]string, len(def.pieces))
	for i, p := range def.pieces {
		p.character = unicode.ToUpper(p.character)
		characters[i] = string(p.character)
	}
	sort.Strings(characters)
	def.alphabet = make(Alphabet, len(characters))
	for i, s := range characters {
		def.alphabet[i] = []rune(s)[0]
	}
	def.initialized = true
}

func SupportedLanguage(language language.Tag) bool {
	_, ok := languageDefinitions[language]
	return ok
}

func getDefinition(language language.Tag) *languageDefinition {
	definition, ok := languageDefinitions[language]
	if !ok {
		panic(fmt.Sprintf("unsupported language %s", language.String()))
	}
	if !definition.initialized {
		definition.init()
	}
	return definition
}

func GetLanguageFileName(language language.Tag) string {
	return getDefinition(language).fileName
}

func GetLanguageTiles(language language.Tag) LanguageTiles {
	return getDefinition(language).pieces
}

func GetLanguageAlphabet(language language.Tag) Alphabet {
	return getDefinition(language).alphabet
}

func (def *languageTile) Character() rune {
	return def.character
}

func (def *languageTile) Count() rune {
	return rune(def.count)
}
func (def *languageTile) Value() rune {
	return rune(def.value)
}
