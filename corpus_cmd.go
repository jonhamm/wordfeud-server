package main

import (
	"flag"
	"fmt"

	"golang.org/x/text/message"
)

func corpusCmd(options *GameOptions, args []string) *CorpusResult {
	result := new(CorpusResult)

	flag := flag.NewFlagSet("exkt", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)

	corpus, err := GetLanguageCorpus(options.language)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	result.Words = make([]string, len(corpus.words))
	for i, w := range corpus.words {
		result.Words[i] = string(w)
	}
	result.WordCount = corpus.wordCount
	result.MinWordLength = corpus.minWordLength
	result.MaxWordLength = corpus.maxWordLength
	result.TotalWordsSize = corpus.totalWordsSize

	p := message.NewPrinter(options.language)
	p.Fprintf(result.logger(), "Number of words  : %d\n", result.WordCount)
	p.Fprintf(result.logger(), "Total words size : %d\n", result.TotalWordsSize)
	p.Fprintf(result.logger(), "Word lengths     : %d .. %d\n", result.MinWordLength, result.MaxWordLength)
	return result.result()
}
