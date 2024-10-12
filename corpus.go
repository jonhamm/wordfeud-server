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
	key           CorpusKey
	words         [][]rune
	wordCount     int
	maxWordLength int
}

type CorpusIndex struct {
	corpus *Corpus
	index  []int
}

func scanWords(f io.Reader) ([][]rune, error) {
	words := make([][]rune, 0)
	ptn := "^[a-zæøå]+$"
	r, err := regexp.Compile(ptn)
	if err != nil {
		return words, err
	}

	s := bufio.NewScanner(f)

	for s.Scan() {
		line := s.Text()
		if !r.MatchString(line) {
			continue
		}
		word := MakeWord(line)
		words = append(words, word)
	}
	sort.Slice(words, func(i int, j int) bool {
		return slices.Compare(words[i], words[j]) < 0
	})

	return words, nil
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

func GetFileCorpus(fileName string) (*Corpus, error) {
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
	corpus, err = NewCorpus(f)
	if err != nil {
		return nil, err
	}
	corpus.key.fileName = fileName
	SetCorpus(corpus)
	return corpus, nil
}

func NewCorpus(content io.Reader) (*Corpus, error) {

	var err error
	corpus := new(Corpus)
	corpus.words, err = scanWords(content)
	corpus.wordCount = len(corpus.words)
	corpus.initStatistics()
	return corpus, err
}

func NewCorpusIndex(corpus *Corpus, index []int) *CorpusIndex {
	return &CorpusIndex{corpus, index}
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

func (corpus *Corpus) PickRandomWord() (wordIndex int, word Word) {
	i := rand.Intn(corpus.wordCount)
	return i, corpus.GetWord(i)
}

func (corpus *Corpus) FindWord(word Word) (wordIndex int, found bool) {
	i := sort.Search(len(corpus.words), func(i int) bool { return slices.Compare(corpus.words[i], word) >= 0 })
	return i, i < len(corpus.words) && slices.Compare(corpus.words[i], word) == 0
}

func (corpus *Corpus) initStatistics() {
	corpus.maxWordLength = 0
	for _, word := range corpus.words {
		if len(word) > corpus.maxWordLength {
			corpus.maxWordLength = len(word)
		}
	}

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
/* func equalWord(lhs Word, rhs Word) bool {
	return slices.Compare(lhs, rhs) == 0
} */

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
