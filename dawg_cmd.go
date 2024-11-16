package main

import (
	"flag"
	"fmt"
	. "wordfeud/context"
	. "wordfeud/corpus"
	. "wordfeud/dawg"

	"golang.org/x/text/message"
)

func dawgCmd(options *GameOptions, args []string) *DawgResult {
	result := new(DawgResult)

	flag := flag.NewFlagSet("exit", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)
	var corpus Corpus
	var content CorpusContent
	var dawg Dawg
	var fileName string
	var err error
	corpus, err = NewCorpus(options.Language)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	fileName = GetLanguageFileName(corpus.Language())
	content, err = corpus.GetFileContent(fileName)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	dawg, err = NewDawg(content, options.Options)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	statistics := DawgStatistics(dawg)
	result.NodeCount = statistics.NodeCount
	result.VertexCount = statistics.VertexCount
	p := message.NewPrinter(options.Language)

	corpusStat := content.Stat()
	p.Fprintf(result.logger(), "Number of words       : %d\n", corpusStat.WordCount)
	p.Fprintf(result.logger(), "Total words size      : %d\n", corpusStat.TotalWordsSize)
	p.Fprintf(result.logger(), "Word lengths          : %d .. %d\n", corpusStat.MinWordLength, corpusStat.MaxWordLength)
	p.Fprintf(result.logger(), "Node count            : %d\n", result.NodeCount)
	p.Fprintf(result.logger(), "Vertex count          : %d\n", result.VertexCount)
	p.Fprintf(result.logger(), "Node and Vertex count : %d   %d%%\n", result.NodeCount+result.VertexCount, ((result.NodeCount+result.VertexCount)*100)/corpusStat.TotalWordsSize)

	return result.result()
}
