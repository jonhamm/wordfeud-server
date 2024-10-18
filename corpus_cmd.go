package main

import (
	"flag"
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func corpusCmd(options *gameOptions, args []string) *CorpusResult {
	result := new(CorpusResult)

	flag := flag.NewFlagSet("solve", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)
	out := options.out
	if options.debug > 0 {
		options.verbose = true
		fmt.Fprintf(out, "options: %+v\n", options)
	}

	corpus, err := GetLangCorpus()
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	result.Words = make([]string, len(corpus.words))
	for i, w := range corpus.words {
		result.Words[i] = string(w)
	}
	result.WordCount = corpus.wordCount
	result.MaxWordLength = corpus.maxWordLength
	result.WordLengthIndex = make([][]int, len(corpus.wordLengthIndex))
	for i, x := range corpus.wordLengthIndex {
		result.WordLengthIndex[i] = x.index
	}
	p := message.NewPrinter(language.Danish)
	p.Fprintf(result.logger(), "Word length frequencies:    \n")
	for i := 1; i <= corpus.MaxWordLength(); i++ {
		p.Fprintf(result.logger(), "   %2d: %8d\n", i, len(corpus.GetWordLengthIndex(i)))
	}
	p.Fprintf(result.logger(), "Number of words: %d\n", result.WordCount)
	return result.result()
}
