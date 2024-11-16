package dawg

type DawgStat struct {
	NodeCount   int
	VertexCount int
}

func DawgStatistics(dawg Dawg) DawgStat {
	var statistics DawgStat
	var nodesVisited = make(map[*node]bool)
	statistics.update(dawg, nodesVisited, dawg.rootNode())
	return statistics
}

func (statistics *DawgStat) update(dawg Dawg, nodesVisited map[*node]bool, node *node) {
	if node == nil {
		return
	}
	if nodesVisited[node] {
		return
	}
	nodesVisited[node] = true
	statistics.NodeCount++
	statistics.VertexCount += len(node.vertices)
	for _, v := range node.vertices {
		statistics.update(dawg, nodesVisited, v.destination)
	}
}
