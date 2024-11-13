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
	var corpus *Corpus
	var content *CorpusContent
	var dawg *Dawg
	var fileName string
	var err error
	corpus, err = NewCorpus(options.language)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	fileName, err = GetLanguageFileName(corpus.language)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	content, err = corpus.GetFileContent(fileName)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	dawg, err = NewDawg(content)
	if err != nil {
		fmt.Println(result.errors(), err.Error())
		return result.result()
	}
	statistics := dawgStatistics(dawg)
	result.NodeCount = statistics.nodeCount
	result.VertexCount = statistics.vertexCount
	p := message.NewPrinter(options.language)

	corpusStat := content.Stat()
	p.Fprintf(result.logger(), "Number of words       : %d\n", corpusStat.wordCount)
	p.Fprintf(result.logger(), "Total words size      : %d\n", corpusStat.totalWordsSize)
	p.Fprintf(result.logger(), "Word lengths          : %d .. %d\n", dawg.corpus.minWordLength, content.maxWordLength)
	p.Fprintf(result.logger(), "Node count            : %d\n", result.NodeCount)
	p.Fprintf(result.logger(), "Vertex count          : %d\n", result.VertexCount)
	p.Fprintf(result.logger(), "Node and Vertex count : %d   %d%%\n", result.NodeCount+result.VertexCount, ((result.NodeCount+result.VertexCount)*100)/content.Stat().totalWordsSize)

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
