package main

import (
	"flag"
	"fmt"

	"golang.org/x/text/message"
)

type DawgStatistics struct {
	nodeCount   int
	vertexCount int
}

func dawgCmd(options *GameOptions, args []string) *DawgResult {
	result := new(DawgResult)

	flag := flag.NewFlagSet("exit", flag.ExitOnError)
	registerGlobalFlags(flag)

	flag.Parse(args)
	out := options.out
	if options.debug > 0 {
		options.verbose = true
		fmt.Fprintf(out, "options: %+v\n", options)
	}
	var corpus *Corpus
	var dawg *Dawg
	var err error
	corpus, err = GetFileCorpus(GetLanguageFileName(options.language), GetLanguageAlphabet((options.language)))
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	dawg, err = NewDawg(corpus)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	statistics := dawgStatistics(dawg)
	result.NodeCount = statistics.nodeCount
	result.VertexCount = statistics.vertexCount
	p := message.NewPrinter(options.language)

	p.Fprintf(result.logger(), "Number of words       : %d\n", dawg.corpus.wordCount)
	p.Fprintf(result.logger(), "Total words size      : %d\n", dawg.corpus.totalWordsSize)
	p.Fprintf(result.logger(), "Word lengths          : %d .. %d\n", dawg.corpus.minWordLength, dawg.corpus.maxWordLength)
	p.Fprintf(result.logger(), "Node count            : %d\n", result.NodeCount)
	p.Fprintf(result.logger(), "Vertex count          : %d\n", result.VertexCount)
	p.Fprintf(result.logger(), "Node and Vertex count : %d   %d%%\n", result.NodeCount+result.VertexCount, ((result.NodeCount+result.VertexCount)*100)/dawg.corpus.totalWordsSize)

	return result.result()
}

func dawgStatistics(dawg *Dawg) DawgStatistics {
	var statistics DawgStatistics
	var nodesVisited = make(map[*Node]bool)
	updateDawgStatistics(&statistics, dawg, nodesVisited, dawg.rootNode)
	return statistics
}

func updateDawgStatistics(statistics *DawgStatistics, dawg *Dawg, nodesVisited map[*Node]bool, node *Node) {
	if node == nil {
		return
	}
	if nodesVisited[node] {
		return
	}
	nodesVisited[node] = true
	statistics.nodeCount++
	statistics.vertexCount += len(node.vertices)
	for _, v := range node.vertices {
		updateDawgStatistics(statistics, dawg, nodesVisited, v.destination)
	}
}
