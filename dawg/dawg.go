package dawg

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	. "wordfeud/context"
	. "wordfeud/corpus"
)

var DAWG_TRACE = false

const dotDir = "/tmp/dot/"

type CRC uint32
type ID uint32

type Dawg interface {
	Corpus() Corpus
	Transition(Letter) DawgState
	Transitions(Word) DawgState
	Match(Word) bool
	FindPrefix(Word) DawgState
	InitialState() DawgState
	rootNode() *node
	options() *Options
}

type _Dawg struct {
	corpus        Corpus
	_options      Options
	_rootNode     *node
	_finalNode    *node
	_initialState DawgState
	nullState     DawgState
	registry      Registry
	nextNodeId    ID
	nextVertexId  ID
}

type vertex struct {
	id          ID
	letter      Letter
	final       bool
	destination *node
}

type vertices []*vertex

type node struct {
	id            ID
	registered    bool
	crc           CRC
	vertices      vertices
	vertexLetters LetterSet // if there is a vertex v in vertices then (vertexLetters & (1 << v.letter))!=0
}

type nodes []*node

type Registry map[CRC]nodes

func NewDawg(content CorpusContent, options Options) (Dawg, error) {
	corpus := content.Corpus()

	dawg := &_Dawg{
		_options:     options,
		corpus:       corpus,
		registry:     make(Registry),
		nextNodeId:   1,
		nextVertexId: 1,
	}
	dawg._rootNode = dawg.newNode()
	dawg._finalNode = dawg.newNode()
	dawg.register(dawg._finalNode)
	dawg.nullState = &_DawgState{dawg: dawg, startNode: nil, vertices: vertices{}}
	dawg._initialState = &_DawgState{dawg: dawg, startNode: dawg._rootNode, vertices: vertices{}}
	err := dawg.addCorpus(content)
	dawg.registry = nil
	if err != nil {
		return nil, err
	}
	return dawg, nil
}

func (dawg *_Dawg) options() *Options {
	return &dawg._options
}

func (dawg *_Dawg) rootNode() *node {
	return dawg._rootNode
}

func (dawg *_Dawg) InitialState() DawgState {
	return dawg._initialState
}

func (dawg *_Dawg) Corpus() Corpus {
	return dawg.corpus
}

func (dawg *_Dawg) trace(format string, a ...any) {
	if DAWG_TRACE {
		fmt.Printf(format, a...)
	}
}

func (dawg *_Dawg) newNode() *node {
	node := &node{
		id:            dawg.nextNodeId,
		registered:    false,
		vertexLetters: 0,
		vertices:      vertices{},
	}
	dawg.nextNodeId++
	return node
}

func (dawg *_Dawg) newVertex(letter Letter, destination *node, final bool) *vertex {
	vertex := &vertex{
		id:          dawg.nextVertexId,
		letter:      letter,
		destination: destination,
		final:       final,
	}
	dawg.nextVertexId++
	return vertex
}

func (dawg *_Dawg) addVertex(node *node, letter Letter, destination *node, final bool) *vertex {
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
	vertex := dawg.newVertex(letter, destination, final)
	node.vertices = append(node.vertices, vertex)
	node.vertexLetters.Set(vertex.letter)
	node.crc = 0 // invalidate crc
	if DAWG_TRACE {
		dawg.printNode(node)
	}
	return vertex
}

func (dawg *_Dawg) lookup(node *node) *node {
	registryNodes := dawg.registry[node.CRC()]
	for _, registryNode := range registryNodes {
		if node.equal(registryNode) {
			return registryNode
		}
	}
	return nil
}

func (dawg *_Dawg) register(node *node) {
	if DAWG_TRACE {
		fmt.Printf("register node#%v\n", node.id)
		dawg.printNode(node)
	}

	n := dawg.lookup(node)
	if n != nil {
		panic(fmt.Sprintf("_Dawg trying to register node:%v allready registered as node:%v (_Dawg.register): ", node.id, n.id))
	}
	if node.registered {
		panic(fmt.Sprintf("_Dawg node:%v is not in registry but node.registered is true (_Dawg.register): ", node.id))
	}
	crc := node.CRC()
	registryNodes := dawg.registry[crc]
	if registryNodes == nil {
		registryNodes = make(nodes, 0, 1)
		dawg.registry[crc] = registryNodes
	}
	dawg.registry[crc] = append(registryNodes, node)
	node.registered = true
}

func (dawg *_Dawg) Transition(letter Letter) DawgState {
	return dawg._initialState.Transition(letter)
}

func (dawg *_Dawg) Transitions(word Word) DawgState {
	return dawg._initialState.Transitions(word)
}

func (dawg *_Dawg) Match(word Word) bool {
	state := dawg.FindPrefix(word)
	if !state.Valid() {
		return false
	}
	if len(word) > len(state.Word()) {
		return false
	}
	return state.Final()
}

func (dawg *_Dawg) FindPrefix(word Word) DawgState {
	state := dawg.InitialState()
	if DAWG_TRACE {
		fmt.Printf("FindPrefix \"%s\" : \n", word.String(dawg.corpus))
		state.Print()
	}

	for _, l := range word {
		nextState := state.Transition(l)
		if !nextState.Valid() {
			break
		}
		state = nextState
	}
	if DAWG_TRACE {
		fmt.Printf("FindPrefix \"%s\" => \n", word.String(dawg.corpus))
		state.Print()
	}
	return state
}

func (dawg *_Dawg) ReplaceOrRegister(node *node) {
	if DAWG_TRACE {
		fmt.Printf("ReplaceOrRegister node#%v\n", node.id)
		dawg.printNode(node)
	}

	lastVertex := node.LastVertex()

	if DAWG_TRACE {
		hasVertices := "has no vertices"
		if lastVertex.destination.HasVertices() {
			hasVertices = "has vertices"
		}
		fmt.Printf("lastvertex#%v('%c').destination: node#%v %s\n", lastVertex.id, dawg.corpus.LetterToRune(lastVertex.letter), node.id, hasVertices)
	}

	if lastVertex.destination.HasVertices() {
		dawg.ReplaceOrRegister(lastVertex.destination)
	}

	if DAWG_TRACE {
		fmt.Printf("-->ReplaceOrRegister node#%v\n", node.id)
	}

	registryNode := dawg.lookup(lastVertex.destination)
	if registryNode != nil {
		lastVertex.destination = registryNode

		if DAWG_TRACE {
			fmt.Printf("lastvertex#%v('%c').destination <= registry node#%v\n", lastVertex.id, dawg.corpus.LetterToRune(lastVertex.letter), registryNode.id)
		}

	} else {
		dawg.register(lastVertex.destination)
	}
}

func (dawg *_Dawg) addCorpus(content CorpusContent) error {
	for i, w := range content.Words() {
		err := dawg.AddWord((w))
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
	dawg.ReplaceOrRegister(dawg._rootNode)
	if DAWG_TRACE {
		dawg.printDot(content.WordCount(), "FINAL")
		dawg.print()
	}
	return nil
}

func (dawg *_Dawg) AddWord(word Word) error {
	if DAWG_TRACE {
		dawg.trace("\n\nAddWord \"%s\"\n\n", word.String(dawg.corpus))
	}
	prefixState := dawg.FindPrefix(word)
	prefixNode := prefixState.lastNode()
	if prefixNode.HasVertices() {
		dawg.ReplaceOrRegister(prefixNode)
	}
	suffix := word[prefixState.WordLength():]
	if len(suffix) > 0 {
		if prefixNode == dawg._finalNode {
			prefixNode = dawg.newNode()
			prefixState.lastVertex().destination = prefixNode
		}
	}
	dawg.AddSuffix(prefixNode, suffix)
	return nil
}

func (dawg *_Dawg) AddSuffix(node *node, suffix Word) {
	if len(suffix) == 0 {
		return
	}
	if DAWG_TRACE {
		fmt.Printf("AddSuffix node#%d \"%s\"\n", node.id, suffix)
	}
	letter := suffix[0]
	suffix = suffix[1:]
	suffixNode := dawg._finalNode
	if len(suffix) > 0 {
		suffixNode = dawg.newNode()
		dawg.AddSuffix(suffixNode, suffix)
	}
	dawg.addVertex(node, letter, suffixNode, suffixNode == dawg._finalNode)
}

func (dawg *_Dawg) fprintfNode(f io.Writer, node *node) {
	if node == nil {
		fmt.Fprint(f, "node#nil\n")
		return
	}
	fmt.Fprintf(f, "node#%v  crc:%v registered:%v vertexLetters:%s\n", node.id, node.crc, node.registered, node.vertexLetters.String(dawg.corpus))
	for i, v := range node.vertices {
		dest := "<nil>"
		if v.destination != nil {
			dest = fmt.Sprintf("node#%v %s", v.destination.id, v.destination.vertexLetters.String(dawg.corpus))
		}
		fmt.Fprintf(f, "  +-- [%v] vertex#%v  letter:'%c' final:%v destination:%s \n", i, v.id, dawg.corpus.LetterToRune(v.letter), v.final, dest)
	}
}

func (dawg *_Dawg) printNode(node *node) {
	dawg.fprintfNode(os.Stdout, node)
}

func (dawg *_Dawg) print() {
	dawg.fprint(os.Stdout)
}

func (dawg *_Dawg) fprint(f io.Writer) {
	fmt.Fprintf(f, "\n\n======== D A W G ========\n\n")
	dawg.fprintfRecurse(f, make(map[*node]bool), dawg._rootNode)
}

func (dawg *_Dawg) fprintfRecurse(f io.Writer, printedNodes map[*node]bool, node *node) {
	if printedNodes[node] {
		return
	}
	dawg.fprintfNode(f, node)
	printedNodes[node] = true
	for _, v := range node.vertices {
		if v.destination != nil {
			dawg.fprintfRecurse(f, printedNodes, v.destination)
		}
	}
}

func (dawg *_Dawg) printDot(seqno int, label string) {
	dotFileName := fmt.Sprintf("%s/%06d_%s.gv", dotDir, seqno, label)
	dotFile, err := os.OpenFile(dotFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("error writing dot file : %s\n%v", dotFileName, err.Error())
		return
	}
	dawg.printDotSubtree(dotFile, label, dawg._rootNode)
	dotFile.Close()
	fmt.Printf("wrote dot file : %s\n", dotFileName)
}

func (dawg *_Dawg) printDotSubtree(f io.Writer, label string, n *node) {
	fmt.Fprint(f, "digraph {\n")
	dawg.printfDotRecurse(f, make(map[*node]bool), false, n)
	dawg.printfDotRecurse(f, make(map[*node]bool), true, n)
	fmt.Fprintf(f, "label=\"\\n\\n%s\"\n", label)
	fmt.Fprint(f, "scale=\"0.5\"\n")
	fmt.Fprint(f, "}\n")

}

func (dawg *_Dawg) printfDotRecurse(f io.Writer, printedNodes map[*node]bool, printVertices bool, node *node) {
	if printedNodes[node] {
		return
	}
	printedNodes[node] = true
	nodeRegistered := ""
	if node.registered {
		nodeRegistered = "+"

	}
	fmt.Fprintf(f, "node [ label=\"%d%s\" ] %d\n", node.id, nodeRegistered, node.id)

	for _, v := range node.vertices {
		if v.destination != nil {
			if printVertices {
				if v.final {
					fmt.Fprintf(f, "%d -> %d [label=\" %c\" arrowhead=\"diamond\"]\n", node.id, v.destination.id, dawg.corpus.LetterToRune(v.letter))
				} else {
					fmt.Fprintf(f, "%d -> %d [label=\" %c\"]\n", node.id, v.destination.id, dawg.corpus.LetterToRune(v.letter))
				}
			}
			dawg.printfDotRecurse(f, printedNodes, printVertices, v.destination)
		}
	}
}

func (node *node) CRC() CRC {
	if node.crc == 0 {
		var e error
		cs := crc32.NewIEEE()
		e = binary.Write(cs, binary.LittleEndian, node.vertexLetters)
		for _, v := range node.vertices {
			if e != nil {
				break
			}
			e = binary.Write(cs, binary.LittleEndian, v.final)
			if e == nil {
				e = binary.Write(cs, binary.LittleEndian, v.destination.CRC())
			}
		}
		if e != nil {
			panic("node failure in ")
		}
		node.crc = CRC(cs.Sum32())
	}
	return node.crc
}

func (node *node) equal(other *node) bool {
	if node == other {
		return true
	}
	if other == nil {
		return false
	}
	if node.vertexLetters != other.vertexLetters {
		return false
	}
	if len(node.vertices) != len(other.vertices) {
		return false
	}
	for i := range node.vertices {
		myVertex := node.vertices[i]
		otherVertex := other.vertices[i]
		if myVertex.letter != otherVertex.letter {
			return false
		}
		if myVertex.final != otherVertex.final {
			return false
		}
		if !myVertex.destination.equal(otherVertex.destination) {
			return false
		}
	}
	return true
}

func (node *node) FindVertex(l Letter) (byte, *vertex) {
	if !node.vertexLetters.Test(l) {
		return byte(len(node.vertices)), nil
	}
	for i, v := range node.vertices {
		if v.letter == l {
			return byte(i), v
		}
	}
	panic("node inconsistent vertexLetters and vertices (node.FindVertex)")
}

func (node *node) HasVertices() bool {
	if (len(node.vertices) == 0) != (node.vertexLetters == 0) {
		panic(fmt.Sprintf("node:%v inconsistent vertexLetters and vertices (node.HasVertices)", node.id))
	}
	return node.vertexLetters != 0
}

func (node *node) IsSameState(otherNode *node) bool {
	if node == otherNode {
		return true
	}
	if node.vertexLetters != otherNode.vertexLetters {
		return false
	}
	if len(node.vertices) != len(otherNode.vertices) {
		panic(fmt.Sprintf("node:%v node:%v inconsistent vertexLetters and vertices (node.IsSameState)", node.id, otherNode.id))
	}
	for i, v := range node.vertices {
		if v.letter != otherNode.vertices[i].letter {
			return false
		}
		if v.destination != otherNode.vertices[i].destination {
			return false
		}
	}
	return true
}

func (node *node) LastVertex() *vertex {
	if (len(node.vertices) == 0) != (node.vertexLetters == 0) {
		panic(fmt.Sprintf("node:%v inconsistent vertexLetters and vertices (node.LastVertexNode)", node.id))
	}
	if node.vertexLetters == 0 {
		return nil
	}
	return node.vertices[len(node.vertices)-1]
}
