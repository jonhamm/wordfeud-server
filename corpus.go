package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/text/language"
)

type Word []rune
type Words [][]rune

type CorpusKey struct {
	fileName string
	language language.Tag
	wordSize int
}
type Corpus struct {
	key                  CorpusKey
	words                [][]rune
	wordCount            int
	allRunes             []rune
	runeFrequencies      map[rune]float64
	wordRuneFrequencies  []float64
	wordVowelFrequencies []float64
	vowels               map[rune]bool
	wordVowelCount       []int8
}

type CorpusIndex struct {
	corpus *Corpus
	index  []int
}

func scanWords(f io.Reader, wordSize int, language language.Tag) ([][]rune, error) {
	words := make([][]rune, 0)
	valid := GetValidLanguageCharacters(language)
	ptn := fmt.Sprintf("^[%s]{%d}$", valid, wordSize)
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
		if len(word) == wordSize {
			words = append(words, word)
		}
	}
	sort.Slice(words, func(i int, j int) bool {
		return slices.Compare(words[i], words[j]) < 0
	})

	return words, nil
}

var corpusCache *lru.Cache

func GetCorpus(fileName string, language language.Tag, wordSize int) *Corpus {
	key := CorpusKey{fileName: fileName, language: language, wordSize: wordSize}
	var corpus *Corpus
	if corpusCache == nil {
		corpusCache, _ = lru.New(10)
	}
	cached, found := corpusCache.Get(key)
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

func GetFileCorpus(fileName string, language language.Tag, wordSize int) (*Corpus, error) {
	fsys := os.DirFS(".")
	corpus := GetCorpus(fileName, language, wordSize)
	if corpus != nil {
		return corpus, nil
	}
	f, err := fsys.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	corpus, err = NewCorpus(language, wordSize, f)
	if err != nil {
		return nil, err
	}
	corpus.key.fileName = fileName
	SetCorpus(corpus)
	return corpus, nil
}

func NewCorpus(language language.Tag, wordSize int, content io.Reader) (*Corpus, error) {

	var err error
	corpus := new(Corpus)
	corpus.key.language = language
	corpus.key.wordSize = wordSize
	corpus.words, err = scanWords(content, wordSize, language)
	corpus.wordCount = len(corpus.words)
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

func (corpus *Corpus) WordSize() int {
	return corpus.key.wordSize
}

func (corpus *Corpus) GetWord(i int) Word {
	if i < 0 || i > corpus.wordCount {
		return make(Word, corpus.WordSize())
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
	if corpus.runeFrequencies != nil {
		return
	}
	corpus.runeFrequencies = make(map[rune]float64)
	corpus.wordRuneFrequencies = make([]float64, corpus.wordCount)
	corpus.vowels = make(map[rune]bool)
	corpus.wordVowelCount = make([]int8, corpus.wordCount)
	corpus.wordVowelFrequencies = make([]float64, corpus.wordCount)
	for _, c := range GetLanguageVowels(corpus.key.language) {
		corpus.vowels[c] = true
	}
	for i, word := range corpus.words {
		runes := make([]rune, corpus.WordSize())
		var vowelCount int8 = 0
		for j, c := range word {
			runes[j] = c
			corpus.runeFrequencies[c] += 1
			if corpus.vowels[c] {
				duplicate := false
				for j > 0 {
					j--
					if runes[j] == c {
						duplicate = true
						break
					}
				}
				if !duplicate {
					vowelCount++
				}
			}
		}
		corpus.wordVowelCount[i] = vowelCount
	}
	total := 0.0
	for _, f := range corpus.runeFrequencies {
		total += f
	}
	for c, f := range corpus.runeFrequencies {
		corpus.runeFrequencies[c] = f / total
	}
	for i, w := range corpus.words {
		f := 0.0
		vf := 0.0
		for j, c := range w {
			for k := j - 1; k >= 0; k-- {
				if w[k] == c {
					f -= 1.0 / float64(len(corpus.runeFrequencies))
					if corpus.vowels[c] {
						vf -= 1.0 / float64(len(corpus.runeFrequencies))
					}
				} else {
					f += corpus.relativeRuneFrequencey(c)
					if corpus.vowels[c] {
						vf += corpus.relativeRuneFrequencey(c)
					}
				}
			}
		}
		corpus.wordRuneFrequencies[i] = f
		corpus.wordVowelFrequencies[i] = vf
	}
	for r := range corpus.runeFrequencies {
		corpus.allRunes = append(corpus.allRunes, r)
	}
	sort.Slice(corpus.allRunes, func(i int, j int) bool {
		return corpus.allRunes[i] < corpus.allRunes[j]
	})
}

func (corpus *Corpus) runes() []rune {
	corpus.initStatistics()
	return corpus.allRunes

}

func (corpus *Corpus) relativeRuneFrequencey(r rune) float64 {
	corpus.initStatistics()
	return corpus.runeFrequencies[r]
}

func (corpus *Corpus) wordRelativeRuneFrequencies(x int) float64 {
	corpus.initStatistics()
	if x >= 0 && x < len(corpus.wordRuneFrequencies) {
		return corpus.wordRuneFrequencies[x]
	}
	return 0.0
}

func (corpus *Corpus) wordRelativeVowelFrequencies(x int) float64 {
	corpus.initStatistics()
	if x >= 0 && x < len(corpus.wordVowelFrequencies) {
		return corpus.wordVowelFrequencies[x]
	}
	return 0.0
}

func (corpusIndex *CorpusIndex) FindIndex(x int) (int, bool) {
	for i, v := range corpusIndex.index {
		if v == x {
			return i, true
		}
	}
	return -1, false
}

func (wordIndex *CorpusIndex) dump() []string {
	corpus := wordIndex.corpus
	result := make([]string, len(wordIndex.index))
	for i, x := range wordIndex.index {
		w := corpus.GetWord(x)
		result[i] = fmt.Sprintf("%d [%d] : '%q'", i, x, w.String())
	}
	return result
}

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
