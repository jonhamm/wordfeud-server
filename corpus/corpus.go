package corpus

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/text/language"
)

type Letter byte
type Word []Letter
type Words []Word

const NoLetter = Letter(0)

type Alphabet []rune

const AlphabetMax = byte(32) // max 32 (0..31) letters in alphabet
// OBS if AlphabetMax is changed, so must the definition below of LetterSet
type LetterSet uint32 // set of Letter - i.e. bitset of 0..31
const NullLetterSet = LetterSet(0)
const AllLetterSet = LetterSet(^NullLetterSet)

type corpusKey struct {
	fileName string
}

type CorpusStat struct {
	WordCount      int
	MinWordLength  int
	MaxWordLength  int
	TotalWordsSize int
}

type Corpus interface {
	RuneToLetter(rune) Letter
	LetterToRune(Letter) rune
	Language() language.Tag
	NewContent(content io.Reader) (CorpusContent, error)
	GetFileContent(string) (CorpusContent, error)
	GetLanguageContent() (CorpusContent, error)
	FirstLetter() Letter
	LastLetter() Letter
	LetterMax() int
	AllLetters() LetterSet
}

type CorpusContent interface {
	Corpus() Corpus
	Words() Words
	WordCount() int
	GetWord(i int) Word
	FindWord(word Word) (wordIndex int, found bool)
	Stat() CorpusStat
}

type corpusData struct {
	key           corpusKey
	language      language.Tag
	alphabet      Alphabet
	allLetters    LetterSet
	letterRune    []rune
	letterMax     Letter
	firstLetter   Letter
	lastLetter    Letter
	runeLetter    map[rune]Letter
	minWordLength int
}

type corpusContent struct {
	corpus        Corpus
	words         Words
	maxWordLength int
	stat          *CorpusStat
}

var corpusCache *lru.Cache

func getCachedCorpus(fileName string) *corpusData {
	var corpus *corpusData
	if corpusCache == nil {
		corpusCache, _ = lru.New(10)
	}
	cached, found := corpusCache.Get(fileName)
	if found {
		corpus = cached.(*corpusData)
	}
	return corpus
}

func setCachedCorpus(corpus *corpusData) {
	if corpusCache == nil {
		corpusCache, _ = lru.New(10)
	}
	corpusCache.Add(corpus.key, corpus)
}

func NewCorpus(lang language.Tag) (Corpus, error) {
	var err error
	corpus := new(corpusData)
	corpus.language = lang
	corpus.alphabet = GetLanguageAlphabet(lang)
	corpus.letterRune = make([]rune, len(corpus.alphabet)+1)
	corpus.runeLetter = make(map[rune]Letter)
	corpus.minWordLength = 2 // scrabble rules : words may not be one letter words
	var n Letter = 0
	for _, r := range corpus.alphabet {
		n++
		if n >= Letter(AlphabetMax) {
			return nil, fmt.Errorf("the alphabet specified has more than %v characters", AlphabetMax)
		}
		corpus.letterRune[n] = r
		corpus.runeLetter[r] = n
		corpus.allLetters.Set(n)
	}
	if n > 0 {
		corpus.firstLetter = 1
		corpus.lastLetter = n
		corpus.letterMax = n + 1
	}
	return corpus, err
}

func (corpus *corpusData) NewContent(content io.Reader) (CorpusContent, error) {
	var err error
	corpusContent := new(corpusContent)
	corpusContent.corpus = corpus
	corpusContent.stat = new(CorpusStat)
	corpusContent.words, err = corpus.scanWords(content)
	for _, w := range corpusContent.words {
		wordLength := len(w)
		if wordLength > corpusContent.maxWordLength {
			corpusContent.maxWordLength = wordLength
		}
		corpusContent.stat.TotalWordsSize += wordLength
	}
	corpusContent.stat.WordCount = len(corpusContent.words)
	corpusContent.stat.MinWordLength = corpus.minWordLength
	corpusContent.stat.MaxWordLength = corpusContent.maxWordLength
	return corpusContent, err
}

func (corpus *corpusData) scanWords(f io.Reader) (Words, error) {
	words := make(Words, 0, 10000)
	var sb strings.Builder
	sb.WriteString("^[")
	for _, l := range corpus.letterRune {
		sb.WriteRune(l)
	}
	sb.WriteString("]+$")
	ptn := sb.String()
	r, err := regexp.Compile(ptn)
	if err != nil {
		return Words{}, err
	}

	s := bufio.NewScanner(f)
	allWords := make(map[string]bool)

	for s.Scan() {
		line := strings.ToUpper(s.Text())
		if !r.MatchString(line) {
			continue
		}
		if allWords[line] {
			continue
		}
		allWords[line] = true

		word := corpus.stringToWord(line)
		if len(word) >= corpus.minWordLength {
			words = append(words, word)
		}
	}

	sort.Slice(words, func(i int, j int) bool {
		return slices.Compare(words[i], words[j]) < 0
	})

	return words, nil
}

func (corpus *corpusData) GetFileContent(fileName string) (CorpusContent, error) {
	fsys := os.DirFS(".")
	var content CorpusContent
	f, err := fsys.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err = corpus.NewContent(f)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (corpus *corpusData) GetLanguageContent() (CorpusContent, error) {
	fileName := GetLanguageFileName(corpus.language)
	return corpus.GetFileContent(fileName)
}

func (content *corpusContent) Corpus() Corpus {
	return content.corpus
}

func (content *corpusContent) Words() Words {
	return content.words
}

func (content *corpusContent) Stat() CorpusStat {
	return *content.stat
}

func (content *corpusContent) WordCount() int {
	return content.stat.WordCount
}

func (content *corpusContent) MaxWordLength() int {
	return content.maxWordLength
}

func (corpus *corpusData) MinWordLength() int {
	return corpus.minWordLength
}

func (corpus *corpusData) LetterMax() int {
	return int(corpus.letterMax)
}

func (corpus *corpusData) FirstLetter() Letter {
	return corpus.firstLetter
}

func (corpus *corpusData) LastLetter() Letter {
	return corpus.firstLetter
}

func (corpus *corpusData) AllLetters() LetterSet {
	return corpus.allLetters
}

func (corpus *corpusData) Language() language.Tag {
	return corpus.language
}

func (corpus *corpusData) RuneToLetter(r rune) Letter {
	return corpus.runeLetter[r]
}

func (corpus *corpusData) LetterToRune(letter Letter) rune {
	if letter < 1 || letter > corpus.letterMax {
		return 0
	}
	return corpus.letterRune[letter]
}

func (content *corpusContent) GetWord(i int) Word {
	if i < 0 || i >= len(content.words) {
		return make(Word, 0)
	}
	return content.words[i]
}

func (content *corpusContent) FindWord(word Word) (wordIndex int, found bool) {
	i := sort.Search(len(content.words), func(i int) bool { return slices.Compare(content.words[i], word) >= 0 })
	return i, i < len(content.words) && slices.Compare(content.words[i], word) == 0
}

func (lhs Word) Equal(rhs Word) bool {
	return slices.Compare(lhs, rhs) == 0
}

func (corpus *corpusData) stringToWord(str string) Word {
	var word Word = make(Word, 0, len(str))
	for _, r := range str {
		l, ok := corpus.runeLetter[r]
		if !ok {
			return Word{}
		}
		word = append(word, l)
	}
	return word
}

func (word Word) String(corpus Corpus) string {
	var str strings.Builder
	for _, c := range word {
		if c == 0 {
			break
		}
		str.WriteRune(corpus.LetterToRune(c))
	}
	return str.String()
}

func (letterSet LetterSet) Test(letter Letter) bool {
	return (letterSet & (1 << letter)) != 0
}

func (letterSet *LetterSet) Set(letter Letter) *LetterSet {
	*letterSet |= LetterSet(1 << letter)
	return letterSet
}

func (letterSet *LetterSet) Unset(letter Letter) *LetterSet {
	*letterSet &^= LetterSet(1 << letter)
	return letterSet
}

func (letterSet LetterSet) String(corpus Corpus) string {
	var s strings.Builder
	var first = true
	s.WriteRune('{')
	for l := corpus.FirstLetter(); l <= corpus.LastLetter(); l++ {
		if letterSet.Test(l) {
			if first {
				first = false
			} else {
				s.WriteRune(',')
			}
			s.WriteRune(rune(corpus.LetterToRune(l)))
		}
	}
	s.WriteRune('}')
	return s.String()
}

func (letter Letter) String(corpus Corpus) string {
	rune := corpus.LetterToRune(letter)
	if rune == 0 {
		return ""
	}
	return string(rune)
}
