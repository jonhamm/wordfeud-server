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
)

type Letter byte
type Word []Letter
type Words []Word

type Alphabet []rune

const AlphabetMax = byte(32) // max 32 (0..31) letters in alphabet
// OBS if AlphabetMax is changed, so must the definition below of LetterSet
type LetterSet uint32 // set of Letter - i.e. bitset of 0..31
const NullLetterSet = LetterSet(0)

type CorpusKey struct {
	fileName string
}

type CorpusIndex struct {
	corpus *Corpus
	index  []int
}

type Corpus struct {
	key            CorpusKey
	alphabet       Alphabet
	allLetters     LetterSet
	letterRune     []rune
	letterMax      Letter
	firstLetter    Letter
	lastLetter     Letter
	runeLetter     map[rune]Letter
	words          Words
	wordCount      int
	minWordLength  int
	maxWordLength  int
	totalWordsSize int
}

var corpusCache *lru.Cache

func GetCorpus(fileName string) *Corpus {
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

func SetCorpus(corpus *Corpus) {
	if corpusCache == nil {
		corpusCache, _ = lru.New(10)
	}
	corpusCache.Add(corpus.key, corpus)
}

func GetFileCorpus(fileName string, alphabet Alphabet) (*Corpus, error) {
	fsys := os.DirFS(".")
	corpus := GetCorpus(fileName)
	if corpus != nil {
		return corpus, nil
	}
	f, err := fsys.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	corpus, err = NewCorpus(f, alphabet)
	if err != nil {
		return nil, err
	}
	corpus.key.fileName = fileName
	SetCorpus(corpus)
	return corpus, nil
}

func NewCorpus(content io.Reader, alphabet Alphabet) (*Corpus, error) {

	var err error
	corpus := new(Corpus)
	corpus.alphabet = alphabet
	corpus.letterRune = make([]rune, len(alphabet)+1)
	corpus.runeLetter = make(map[rune]Letter)
	corpus.minWordLength = 2 // scrabble rules : words may not be one letter words
	var n Letter = 0
	for _, r := range alphabet {
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
	corpus.words, err = corpus.scanWords(content)
	corpus.wordCount = len(corpus.words)
	return corpus, err
}

func NewCorpusIndex(corpus *Corpus, index []int) *CorpusIndex {
	return &CorpusIndex{corpus, index}
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
		return words, err
	}

	s := bufio.NewScanner(f)
	corpus.maxWordLength = 0

	for s.Scan() {
		line := strings.ToUpper(s.Text())
		if !r.MatchString(line) {
			continue
		}
		word := corpus.StringToWord(line)
		if len(word) >= corpus.minWordLength {
			words = append(words, word)
			wordLength := len(word)
			if wordLength > corpus.maxWordLength {
				corpus.maxWordLength = wordLength
			}
			corpus.totalWordsSize += wordLength
		}
	}
	sort.Slice(words, func(i int, j int) bool {
		return slices.Compare(words[i], words[j]) < 0
	})

	return words, nil
}

func (corpus *Corpus) WordList() Words {
	return corpus.words
}

func (corpus *Corpus) WordCount() int {
	return corpus.wordCount
}

func (corpus *Corpus) MaxWordLength() int {
	return corpus.maxWordLength
}

func (corpus *Corpus) GetWord(i int) Word {
	if i < 0 || i > corpus.wordCount {
		return make(Word, 0)
	}
	return corpus.words[i]
}

func (corpus *Corpus) FindWord(word Word) (wordIndex int, found bool) {
	i := sort.Search(len(corpus.words), func(i int) bool { return slices.Compare(corpus.words[i], word) >= 0 })
	return i, i < len(corpus.words) && slices.Compare(corpus.words[i], word) == 0
}

func (corpusIndex *CorpusIndex) FindIndex(x int) (int, bool) {
	for i, v := range corpusIndex.index {
		if v == x {
			return i, true
		}
	}
	return -1, false
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

func (index *CorpusIndex) Contains(x int) bool {
	return slices.Contains(index.index, x)
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
