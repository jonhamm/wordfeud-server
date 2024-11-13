package main

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

type CorpusKey struct {
	fileName string
}

type CorpusStat struct {
	wordCount      int
	totalWordsSize int
}

type Corpus struct {
	key           CorpusKey
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

type CorpusContent struct {
	corpus        *Corpus
	words         Words
	maxWordLength int
	stat          CorpusStat
}

var corpusCache *lru.Cache

func getCachedCorpus(fileName string) *Corpus {
	var corpus *Corpus
	if corpusCache == nil {
		corpusCache, _ = lru.New(10)
	}
	cached, found := corpusCache.Get(fileName)
	if found {
		corpus = cached.(*Corpus)
	}
	return corpus
}

func setCachedCorpus(corpus *Corpus) {
	if corpusCache == nil {
		corpusCache, _ = lru.New(10)
	}
	corpusCache.Add(corpus.key, corpus)
}

func NewCorpus(lang language.Tag) (*Corpus, error) {
	var err error
	corpus := new(Corpus)
	corpus.language = lang
	corpus.alphabet, err = GetLanguageAlphabet(lang)
	if err != nil {
		return nil, err
	}
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
		corpus.allLetters.set(n)
	}
	if n > 0 {
		corpus.firstLetter = 1
		corpus.lastLetter = n
		corpus.letterMax = n + 1
	}
	return corpus, err
}

func (corpus *Corpus) NewContent(content io.Reader) (*CorpusContent, error) {
	var err error
	corpusContent := new(CorpusContent)
	corpusContent.corpus = corpus
	corpusContent.words, err = corpus.scanWords(content)
	for _, w := range corpusContent.words {
		wordLength := len(w)
		if wordLength > corpusContent.maxWordLength {
			corpusContent.maxWordLength = wordLength
		}
		corpusContent.stat.totalWordsSize += wordLength
	}
	corpusContent.stat.wordCount = len(corpusContent.words)
	return corpusContent, err
}

func (corpus *Corpus) scanWords(f io.Reader) (Words, error) {
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

	for s.Scan() {
		line := strings.ToUpper(s.Text())
		if !r.MatchString(line) {
			continue
		}
		word := corpus.StringToWord(line)
		if len(word) >= corpus.minWordLength {
			words = append(words, word)
		}
	}

	sort.Slice(words, func(i int, j int) bool {
		return slices.Compare(words[i], words[j]) < 0
	})

	return words, nil
}

func (corpus *Corpus) GetFileContent(fileName string) (*CorpusContent, error) {
	fsys := os.DirFS(".")
	var content *CorpusContent
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

func (corpus *Corpus) GetLanguageContent() (*CorpusContent, error) {
	fileName, err := GetLanguageFileName(corpus.language)
	if err != nil {
		return nil, fmt.Errorf("unsupported language %s", corpus.language.String())
	}
	return corpus.GetFileContent(fileName)
}

func (content *CorpusContent) WordList() Words {
	return content.words
}

func (content *CorpusContent) Stat() CorpusStat {
	return content.stat
}

func (content *CorpusContent) WordCount() int {
	return content.stat.wordCount
}

func (content *CorpusContent) MaxWordLength() int {
	return content.maxWordLength
}

func (corpus *Corpus) MinWordLength() int {
	return corpus.minWordLength
}

func (content *CorpusContent) GetWord(i int) Word {
	if i < 0 || i >= len(content.words) {
		return make(Word, 0)
	}
	return content.words[i]
}

func (content *CorpusContent) FindWord(word Word) (wordIndex int, found bool) {
	i := sort.Search(len(content.words), func(i int) bool { return slices.Compare(content.words[i], word) >= 0 })
	return i, i < len(content.words) && slices.Compare(content.words[i], word) == 0
}

func (lhs Word) equal(rhs Word) bool {
	return slices.Compare(lhs, rhs) == 0
}

func (corpus *Corpus) StringToWord(str string) Word {
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

func (corpus *Corpus) String(letterSet *LetterSet) string {
	return letterSet.String((corpus))
}

func (word Word) String(corpus *Corpus) string {
	var str strings.Builder
	for _, c := range word {
		if c == 0 {
			break
		}
		str.WriteRune(corpus.letterRune[c])
	}
	return str.String()
}

func (letterSet LetterSet) test(letter Letter) bool {
	return (letterSet & (1 << letter)) != 0
}

func (letterSet *LetterSet) set(letter Letter) *LetterSet {
	*letterSet |= LetterSet(1 << letter)
	return letterSet
}

func (letterSet *LetterSet) unset(letter Letter) *LetterSet {
	*letterSet &^= LetterSet(1 << letter)
	return letterSet
}

func (letterSet *LetterSet) String(corpus *Corpus) string {
	var s strings.Builder
	var first = true
	s.WriteRune('{')
	for l := corpus.firstLetter; l <= corpus.lastLetter; l++ {
		if letterSet.test(l) {
			if first {
				first = false
			} else {
				s.WriteRune(',')
			}
			s.WriteRune(rune(corpus.letterRune[l]))
		}
	}
	s.WriteRune('}')
	return s.String()
}

func (letter Letter) String(corpus *Corpus) string {
	if letter == 0 {
		return ""
	}
	return string(corpus.letterRune[letter])
}

func (corpus *Corpus) WordToString(word Word) string {
	var sb strings.Builder
	sb.WriteString("'")
	sb.WriteString(word.String(corpus))
	sb.WriteString("'")

	return sb.String()
}
