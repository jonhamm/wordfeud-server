package dawg

import (
	"fmt"
	"os"
	. "wordfeud/corpus"
)

type crc uint32
type vertexID uint32
type nodeID uint32

type Registry map[crc]nodes

type dawgBuilder struct {
	dawg         *_Dawg
	registry     Registry
	nextNodeId   nodeID
	nextVertexId vertexID
}

func (builder *dawgBuilder) addCorpus(content CorpusContent) error {
	dawg := builder.dawg
	for i, w := range content.Words() {
		err := builder.addWord((w))
		if err != nil {
			return err
		}
		if DAWG_TRACE {
			if i == 0 {
				err = os.RemoveAll(dotDir)
				if err != nil {
					return err
				}
				err = os.MkdirAll(dotDir, 0744)
				if err != nil {
					return err
				}
			}
			dawg.printDot(i, w.String(dawg.corpus))
		}
	}
	builder.replaceOrRegister(dawg._rootNode)
	if DAWG_TRACE {
		dawg.printDot(content.WordCount(), "FINAL")
		dawg.print()
	}
	return nil
}

func (builder *dawgBuilder) addWord(word Word) error {
	dawg := builder.dawg
	if DAWG_TRACE {
		dawg.trace("\n\nAddWord \"%s\"\n\n", word.String(dawg.corpus))
	}
	prefixState := dawg.FindPrefix(word)
	prefixNode := prefixState.lastNode()
	if prefixNode.hasVertices() {
		builder.replaceOrRegister(prefixNode)
	}
	suffix := word[prefixState.WordLength():]
	if len(suffix) > 0 {
		if prefixNode == dawg._finalNode {
			prefixNode = builder.newNode()
			prefixState.lastVertex().destination = prefixNode
		}
	}
	builder.addSuffix(prefixNode, suffix)
	return nil
}

func (builder *dawgBuilder) addSuffix(node *node, suffix Word) {
	dawg := builder.dawg
	if len(suffix) == 0 {
		return
	}
	if DAWG_TRACE {
		fmt.Printf("addSuffix node#%d \"%s\"\n", node.id, suffix)
	}
	letter := suffix[0]
	suffix = suffix[1:]
	suffixNode := dawg._finalNode
	if len(suffix) > 0 {
		suffixNode = builder.newNode()
		builder.addSuffix(suffixNode, suffix)
	}
	builder.addVertex(node, letter, suffixNode, suffixNode == dawg._finalNode)
}

func (builder *dawgBuilder) replaceOrRegister(node *node) {
	dawg := builder.dawg
	if DAWG_TRACE {
		fmt.Printf("replaceOrRegister node#%v\n", node.id)
		dawg.printNode(node)
	}

	lastVertex := node.lastVertex()

	if DAWG_TRACE {
		hasVertices := "has no vertices"
		if lastVertex.destination.hasVertices() {
			hasVertices = "has vertices"
		}
		fmt.Printf("lastvertex#%v('%c').destination: node#%v %s\n", lastVertex.id, dawg.corpus.LetterToRune(lastVertex.letter), node.id, hasVertices)
	}

	if lastVertex.destination.hasVertices() {
		builder.replaceOrRegister(lastVertex.destination)
	}

	if DAWG_TRACE {
		fmt.Printf("-->replaceOrRegister node#%v\n", node.id)
	}

	registryNode := builder.lookup(lastVertex.destination)
	if registryNode != nil {
		lastVertex.destination = registryNode

		if DAWG_TRACE {
			fmt.Printf("lastvertex#%v('%c').destination <= registry node#%v\n", lastVertex.id, dawg.corpus.LetterToRune(lastVertex.letter), registryNode.id)
		}

	} else {
		builder.register(lastVertex.destination)
	}
}

func (builder *dawgBuilder) addVertex(node *node, letter Letter, destination *node, final bool) *vertex {
	dawg := builder.dawg
	if DAWG_TRACE {
		fmt.Printf("addVertex node#%v letter:'%c' destination:node%v final:%v\n", node.id, dawg.corpus.LetterToRune(letter), destination.id, final)
		dawg.printNode(node)
	}
	if node.vertexLetters.Test(letter) {
		panic(fmt.Sprintf("node:%v trying to add vertex with an allready present letter ('%c') (node.addVertex)", node.id, dawg.corpus.LetterToRune(letter)))
	}
	if node.registered {
		panic(fmt.Sprintf("node:%v trying to add vertex with letter ('%c') to registered node (node.addVertex)", node.id, dawg.corpus.LetterToRune(letter)))
	}
	vertex := builder.newVertex(letter, destination, final)
	node.vertices = append(node.vertices, vertex)
	node.vertexLetters.Set(vertex.letter)
	node._crc = 0 // invalidate crc
	if DAWG_TRACE {
		dawg.printNode(node)
	}
	return vertex
}

func (builder *dawgBuilder) lookup(node *node) *node {
	registryNodes := builder.registry[node.crc()]
	for _, registryNode := range registryNodes {
		if node.equal(registryNode) {
			return registryNode
		}
	}
	return nil
}

func (builder *dawgBuilder) register(node *node) {
	dawg := builder.dawg
	if DAWG_TRACE {
		fmt.Printf("register node#%v\n", node.id)
		dawg.printNode(node)
	}

	n := builder.lookup(node)
	if n != nil {
		panic(fmt.Sprintf("_Dawg trying to register node:%v allready registered as node:%v (_builder.register): ", node.id, n.id))
	}
	if node.registered {
		panic(fmt.Sprintf("_Dawg node:%v is not in registry but node.registered is true (_builder.register): ", node.id))
	}
	crc := node.crc()
	registryNodes := builder.registry[crc]
	if registryNodes == nil {
		registryNodes = make(nodes, 0, 1)
		builder.registry[crc] = registryNodes
	}
	builder.registry[crc] = append(registryNodes, node)
	node.registered = true
}

func (builder *dawgBuilder) newNode() *node {
	node := &node{
		id:            builder.nextNodeId,
		registered:    false,
		vertexLetters: 0,
		vertices:      vertices{},
	}
	builder.nextNodeId++
	return node
}

func (builder *dawgBuilder) newVertex(letter Letter, destination *node, final bool) *vertex {
	vertex := &vertex{
		id:          builder.nextVertexId,
		letter:      letter,
		destination: destination,
		final:       final,
	}
	builder.nextVertexId++
	return vertex
}
