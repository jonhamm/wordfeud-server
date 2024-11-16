package main

import (
	"flag"
	"fmt"
	. "wordfeud/context"
	. "wordfeud/corpus"

	"golang.org/x/text/message"
)

func corpusCmd(options *GameOptions, args []string) *CorpusResult {
	result := new(CorpusResult)

	flag := flag.NewFlagSet("exkt", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)

	corpus, err := NewCorpus(options.Language)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	content, err := corpus.GetLanguageContent()
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	result.Words = make([]string, len(content.Words()))
	for i, w := range content.Words() {
		result.Words[i] = string(w)
	}
	corpusStat := content.Stat()
	result.WordCount = corpusStat.WordCount
	result.MinWordLength = corpusStat.MinWordLength
	result.MaxWordLength = corpusStat.MaxWordLength
	result.TotalWordsSize = corpusStat.TotalWordsSize

	p := message.NewPrinter(options.Language)
	p.Fprintf(result.logger(), "Number of words  : %d\n", result.WordCount)
	p.Fprintf(result.logger(), "Total words size : %d\n", result.TotalWordsSize)
	p.Fprintf(result.logger(), "Word lengths     : %d .. %d\n", result.MinWordLength, result.MaxWordLength)
	return result.result()
}
