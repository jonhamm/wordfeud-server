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
	result.Words = corpus.words
	result.WordCount = corpus.wordCount
	result.MaxWordLength = corpus.maxWordLength
	p := message.NewPrinter(language.Danish)
	p.Fprintf(result.logger(), "Number of words: %d\n", result.WordCount)
	p.Fprintf(out, "Longest word:    %d\n", result.MaxWordLength)
	return result.result()
}
