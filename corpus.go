package main

import (
	"bufio"
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

type CorpusKey struct {
	fileName string
}

type CorpusIndex struct {
	corpus *Corpus
	index  []int
}

type Corpus struct {
	key             CorpusKey
	letterRune      []rune
	letterMax       Letter
	firstLetter     Letter
	lastLetter      Letter
	runeLetter      map[rune]Letter
	words           Words
	wordCount       int
	maxWordLength   int
	wordLengthIndex [] /*wordlength*/ *CorpusIndex
	letterPosIndex  [] /*Letter*/ [] /*letterPos*/ *CorpusIndex
	pieces          LanguagePieces
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

func NewCorpus(content io.Reader, pieces LanguagePieces) (*Corpus, error) {

	var err error
	corpus := new(Corpus)
	corpus.pieces = pieces
	corpus.letterRune = make([]rune, len(pieces)+1)
	corpus.runeLetter = make(map[rune]Letter)
	var n Letter = 0
	for _, p := range pieces {
		n++
		corpus.letterRune[n] = p.character
		corpus.runeLetter[p.character] = n
	}
	if n > 0 {
		corpus.firstLetter = 1
		corpus.lastLetter = n
		corpus.letterMax = n + 1
	}
	corpus.words, err = corpus.scanWords(content)
	corpus.wordCount = len(corpus.words)
	corpus.initStatistics()
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
		line := s.Text()
		if !r.MatchString(line) {
			continue
		}
		word := corpus.MakeWord(line)
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
	corpus.wordLengthIndex = make([]*CorpusIndex, corpus.maxWordLength+1)
	for i := range corpus.wordLengthIndex {
		corpus.wordLengthIndex[i] = NewCorpusIndex(corpus, []int{})
	}

	corpus.letterPosIndex = make([][]*CorpusIndex, corpus.letterMax)
	for l := Letter(0); l < corpus.letterMax; l++ {
		corpus.letterPosIndex[l] = make([]*CorpusIndex, 0)
	}
	for i, word := range corpus.words {
		length := len(word)
		corpus.wordLengthIndex[length].index = append(corpus.wordLengthIndex[length].index, i)

		for p, l := range word {
			index := corpus.letterPosIndex[l]
			if p >= len(index) {
				index = slices.Grow(index, p+1-len(index))
				for n := len(index); n <= p; n++ {
					index = append(index, nil)
				}
			}
			if index[p] == nil {
				index[p] = NewCorpusIndex(corpus, []int{})
			}
			index[p].index = append(index[p].index, i)
			corpus.letterPosIndex[l] = index
		}

	}

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

func (corpus *Corpus) GetWordLengthIndex(length int) []int {
	if length < 1 || length >= len(corpus.wordLengthIndex) {
		return make([]int, 0)
	}
	return corpus.wordLengthIndex[length].index
}
func (corpus *Corpus) GetPositionIndex(character Letter, position int) *CorpusIndex {
	index := corpus.letterPosIndex[character]
	if position < 0 || position >= len(index) || index[position] == nil {
		return &CorpusIndex{corpus, make([]int, 0)}
	}
	return index[position]
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

func (corpus *Corpus) MakeWord(str string) Word {
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
