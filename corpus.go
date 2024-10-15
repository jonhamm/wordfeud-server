package main

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	lru "github.com/hashicorp/golang-lru"
)

type Word []rune
type Words [][]rune

type CorpusKey struct {
	fileName string
}
type Corpus struct {
	key             CorpusKey
	words           Words
	wordCount       int
	maxWordLength   int
	wordLengthIndex []CorpusIndex
	pieces          LanguagePieces
}

type CorpusIndex struct {
	corpus *Corpus
	index  []int
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

func GetFileCorpus(fileName string, pieces LanguagePieces) (*Corpus, error) {
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
	corpus, err = NewCorpus(f, pieces)
	if err != nil {
		return nil, err
	}
	corpus.key.fileName = fileName
	SetCorpus(corpus)
	return corpus, nil
}

func NewCorpus(content io.Reader, pices LanguagePieces) (*Corpus, error) {

	var err error
	corpus := new(Corpus)
	corpus.pieces = pices
	corpus.words, err = corpus.scanWords(content)
	corpus.wordCount = len(corpus.words)
	corpus.initStatistics()
	return corpus, err
}

func NewCorpusIndex(corpus *Corpus, index []int) *CorpusIndex {
	return &CorpusIndex{corpus, index}
}

func (corpus *Corpus) scanWords(f io.Reader) ([][]rune, error) {
	words := make([][]rune, 0, 10000)
	var sb strings.Builder
	sb.WriteString("^[")
	for _, r := range corpus.pieces {

		sb.WriteRune(r.character)
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
		line := s.Text()
		if !r.MatchString(line) {
			continue
		}
		word := MakeWord(line)
		words = append(words, word)
		if len(word) > corpus.maxWordLength {
			corpus.maxWordLength = len(word)

		}
	}
	sort.Slice(words, func(i int, j int) bool {
		return slices.Compare(words[i], words[j]) < 0
	})

	return words, nil
}

func (corpus *Corpus) initStatistics() {
	corpus.wordLengthIndex = make([]CorpusIndex, corpus.maxWordLength+1)

	for i, word := range corpus.words {
		length := len(word)
		corpus.wordLengthIndex[length].index = append(corpus.wordLengthIndex[length].index, i)
	}
}

func (corpus *Corpus) WordList() [][]rune {
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

func (corpus *Corpus) GetWordLengthIndex(length int) []int {
	if length < 1 || length >= len(corpus.wordLengthIndex) {
		return make([]int, 0)
	}
	return corpus.wordLengthIndex[length].index
}

func (corpus *Corpus) PickRandomWord() (wordIndex int, word Word) {
	i := rand.Intn(corpus.wordCount)
	return i, corpus.GetWord(i)
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

/* func (wordIndex *CorpusIndex) dump() []string {
	corpus := wordIndex.corpus
	result := make([]string, len(wordIndex.index))
	for i, x := range wordIndex.index {
		w := corpus.GetWord(x)
		result[i] = fmt.Sprintf("%d [%d] : '%q'", i, x, w.String())
	}
	return result
}
*/

func equalWord(lhs Word, rhs Word) bool {
	return slices.Compare(lhs, rhs) == 0
}

func MakeWord(str string) Word {
	var word Word = make([]rune, 0, len(str))
	for _, r := range str {
		word = append(word, r)
	}
	return word
}

func (word Word) String() string {
	var str strings.Builder
	for _, c := range word {
		if c == 0 {
			break
		}
		str.WriteRune(c)
	}
	return str.String()
}
