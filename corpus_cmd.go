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

	corpus, err := NewCorpus(options.language)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	content, err := corpus.GetLanguageContent()
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	result.Words = make([]string, len(content.words))
	for i, w := range content.words {
		result.Words[i] = string(w)
	}
	corpusStat := content.Stat()
	result.WordCount = corpusStat.wordCount
	result.MinWordLength = corpus.minWordLength
	result.MaxWordLength = content.maxWordLength
	result.TotalWordsSize = corpusStat.totalWordsSize

	p := message.NewPrinter(options.language)
	p.Fprintf(result.logger(), "Number of words  : %d\n", result.WordCount)
	p.Fprintf(result.logger(), "Total words size : %d\n", result.TotalWordsSize)
	p.Fprintf(result.logger(), "Word lengths     : %d .. %d\n", result.MinWordLength, result.MaxWordLength)
	return result.result()
}
