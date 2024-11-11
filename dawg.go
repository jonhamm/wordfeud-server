package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"slices"
)

var DAWG_TRACE = false

const dotDir = "/tmp/dot/"

type CRC uint32
type ID uint32

type Vertex struct {
	id          ID
	letter      Letter
	final       bool
	destination *Node
}

type Vertices []*Vertex

type Node struct {
	id            ID
	registered    bool
	crc           CRC
	vertices      Vertices
	vertexLetters LetterSet // if there is a vertex v in vertices then (vertexLetters & (1 << v.letter))!=0
}

type Nodes []*Node

type DawgState struct {
	startNode *Node
	vertices  Vertices
}

var NullState = DawgState{}

type Registry map[CRC]Nodes

type Dawg struct {
	corpus       *Corpus
	rootNode     *Node
	finalNode    *Node
	initialState DawgState
	registry     Registry
	nextNodeId   ID
	nextVertexId ID
}

func (dawg *Dawg) trace(format string, a ...any) {
	if DAWG_TRACE {
		fmt.Printf(format, a...)
	}
}

func (state *DawgState) Word() Word {
	word := make(Word, len(state.vertices))
	for i, v := range state.vertices {
		word[i] = v.letter
	}
	return word
}

func (state *DawgState) WordLength() int {
	return len(state.vertices)
}

func (node *Node) CRC() CRC {
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
			panic("Node failure in ")
		}
		node.crc = CRC(cs.Sum32())
	}
	return node.crc
}

func (node *Node) equal(other *Node) bool {
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

func (node *Node) FindVertex(l Letter) (byte, *Vertex) {
	if !node.vertexLetters.test(l) {
		return byte(len(node.vertices)), nil
	}
	for i, v := range node.vertices {
		if v.letter == l {
			return byte(i), v
		}
	}
	panic("Node inconsistent vertexLetters and vertices (Node.FindVertex)")
}

func (node *Node) HasVertices() bool {
	if (len(node.vertices) == 0) != (node.vertexLetters == 0) {
		panic(fmt.Sprintf("Node:%v inconsistent vertexLetters and vertices (Node.HasVertices)", node.id))
	}
	return node.vertexLetters != 0
}

func (node *Node) IsSameState(otherNode *Node) bool {
	if node == otherNode {
		return true
	}
	if node.vertexLetters != otherNode.vertexLetters {
		return false
	}
	if len(node.vertices) != len(otherNode.vertices) {
		panic(fmt.Sprintf("Node:%v Node:%v inconsistent vertexLetters and vertices (Node.IsSameState)", node.id, otherNode.id))
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

func (node *Node) LastVertex() *Vertex {
	if (len(node.vertices) == 0) != (node.vertexLetters == 0) {
		panic(fmt.Sprintf("Node:%v inconsistent vertexLetters and vertices (Node.LastVertexNode)", node.id))
	}
	if node.vertexLetters == 0 {
		return nil
	}
	return node.vertices[len(node.vertices)-1]
}

func (state *DawgState) LastVertex() *Vertex {
	if state.startNode == nil {
		return nil
	}
	n := len(state.vertices)
	if n == 0 {
		return nil
	}
	return state.vertices[n-1]
}

func (state *DawgState) LastNode() *Node {
	lastVertex := state.LastVertex()
	if lastVertex == nil {
		return state.startNode
	}

	return lastVertex.destination
}

func NewDawg(corpus *Corpus) (*Dawg, error) {

	dawg := &Dawg{
		corpus:       corpus,
		registry:     make(Registry),
		nextNodeId:   1,
		nextVertexId: 1,
	}

	dawg.rootNode = dawg.NewNode()
	dawg.finalNode = dawg.NewNode()
	dawg.Register(dawg.finalNode)
	dawg.initialState = DawgState{startNode: dawg.rootNode, vertices: Vertices{}}
	err := dawg.AddCorpus(corpus)
	if err != nil {
		return nil, err
	}
	return dawg, nil
}

func (dawg *Dawg) NewNode() *Node {
	node := &Node{
		id:            dawg.nextNodeId,
		registered:    false,
		vertexLetters: 0,
		vertices:      Vertices{},
	}
	dawg.nextNodeId++
	return node
}

func (dawg *Dawg) NewVertex(letter Letter, destination *Node, final bool) *Vertex {
	vertex := &Vertex{
		id:          dawg.nextVertexId,
		letter:      letter,
		destination: destination,
		final:       final,
	}
	dawg.nextVertexId++
	return vertex
}

func (dawg *Dawg) AddVertex(node *Node, letter Letter, destination *Node, final bool) *Vertex {
	if DAWG_TRACE {
		fmt.Printf("AddVertex node#%v letter:'%c' destination:node%v final:%v\n", node.id, dawg.corpus.letterRune[letter], destination.id, final)
		dawg.printNode(node)
	}
	if node.vertexLetters.test(letter) {
		panic(fmt.Sprintf("Node:%v trying to add vertex with an allready present letter ('%c') (Node.AddVertex)", node.id, dawg.corpus.letterRune[letter]))
	}
	if node.registered {
		panic(fmt.Sprintf("Node:%v trying to add vertex with letter ('%c') to registered node (Node.AddVertex)", node.id, dawg.corpus.letterRune[letter]))
	}
	vertex := dawg.NewVertex(letter, destination, final)
	node.vertices = append(node.vertices, vertex)
	node.vertexLetters.set(vertex.letter)
	node.crc = 0 // invalidate crc
	if DAWG_TRACE {
		dawg.printNode(node)
	}
	return vertex
}

func (dawg *Dawg) Lookup(node *Node) *Node {
	registryNodes := dawg.registry[node.CRC()]
	for _, registryNode := range registryNodes {
		if node.equal(registryNode) {
			return registryNode
		}
	}
	return nil
}

func (dawg *Dawg) Register(node *Node) {
	if DAWG_TRACE {
		fmt.Printf("Register node#%v\n", node.id)
		dawg.printNode(node)
	}

	n := dawg.Lookup(node)
	if n != nil {
		panic(fmt.Sprintf("Dawg trying to register node:%v allready registered as node:%v (Dawg.Register): ", node.id, n.id))
	}
	if node.registered {
		panic(fmt.Sprintf("Dawg node:%v is not in registry but node.registered is true (Dawg.Register): ", node.id))
	}
	crc := node.CRC()
	registryNodes := dawg.registry[crc]
	if registryNodes == nil {
		registryNodes = make(Nodes, 0, 1)
		dawg.registry[crc] = registryNodes
	}
	dawg.registry[crc] = append(registryNodes, node)
	node.registered = true
}

func (dawg *Dawg) Transition(state DawgState, letter Letter) DawgState {
	if DAWG_TRACE {
		fmt.Printf("Transition '%c' on state: \n", dawg.corpus.letterRune[letter])
		dawg.printState(state)
	}

	if state.startNode == nil {
		if DAWG_TRACE {
			fmt.Printf("Transition '%c' on null state => NullState\n", dawg.corpus.letterRune[letter])
		}
		return NullState
	}
	node := state.LastNode()
	if node == nil {
		if DAWG_TRACE {
			fmt.Printf("Transition '%c' on nil destination => NullState\n", dawg.corpus.letterRune[letter])
		}
		return NullState
	}

	_, v := node.FindVertex(letter)
	if v == nil {
		if DAWG_TRACE {
			fmt.Printf("vertext for letter '%c' not found in node#%v  => NullState\n", dawg.corpus.letterRune[letter], node.id)
			dawg.printNode(node)
		}
		return NullState
	}

	transitionState := DawgState{startNode: state.startNode, vertices: slices.Concat(state.vertices, Vertices{v})}

	if DAWG_TRACE {
		fmt.Printf("Transition '%c' in node#%v  => vertex#%v node#%v final:%v word:\"%s\"\n",
			dawg.corpus.letterRune[letter], node.id, v.id, v.destination.id, v.final, transitionState.Word().String(dawg.corpus))
		dawg.printState(transitionState)
	}

	return transitionState
}

func (dawg *Dawg) Transitions(state DawgState, word Word) DawgState {
	if len(word) == 0 {
		return state
	}
	state = dawg.Transition(state, word[0])
	return dawg.Transitions(state, word[1:])
}

func (dawg *Dawg) Match(word Word) bool {
	state := dawg.FindPrefix(word)
	if state.startNode == nil {
		return false
	}
	if len(word) > len(state.Word()) {
		return false
	}
	v := state.LastVertex()
	return v.final
}

func (dawg *Dawg) FindPrefix(word Word) DawgState {
	state := dawg.initialState
	if DAWG_TRACE {
		fmt.Printf("FindPrefix \"%s\" : \n", word.String(dawg.corpus))
		dawg.printState(state)
	}

	for _, l := range word {
		nextState := dawg.Transition(state, l)
		if nextState.startNode == nil {
			break
		}
		state = nextState
	}
	if DAWG_TRACE {
		fmt.Printf("FindPrefix \"%s\" => \n", word.String(dawg.corpus))
		dawg.printState(state)
	}
	return state
}

func (dawg *Dawg) ReplaceOrRegister(node *Node) {
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
		fmt.Printf("lastvertex#%v('%c').destination: node#%v %s\n", lastVertex.id, dawg.corpus.letterRune[lastVertex.letter], node.id, hasVertices)
	}

	if lastVertex.destination.HasVertices() {
		dawg.ReplaceOrRegister(lastVertex.destination)
	}

	if DAWG_TRACE {
		fmt.Printf("-->ReplaceOrRegister node#%v\n", node.id)
	}

	registryNode := dawg.Lookup(lastVertex.destination)
	if registryNode != nil {
		lastVertex.destination = registryNode

		if DAWG_TRACE {
			fmt.Printf("lastvertex#%v('%c').destination <= registry node#%v\n", lastVertex.id, dawg.corpus.letterRune[lastVertex.letter], registryNode.id)
		}

	} else {
		dawg.Register(lastVertex.destination)
	}
}

func (dawg *Dawg) AddCorpus(corpus *Corpus) error {
	for i, w := range corpus.words {
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
	dawg.ReplaceOrRegister(dawg.rootNode)
	if DAWG_TRACE {
		dawg.printDot(corpus.WordCount(), "FINAL")
		dawg.print()
	}
	return nil
}

func (dawg *Dawg) AddWord(word Word) error {
	if DAWG_TRACE {
		dawg.trace("\n\nAddWord \"%s\"\n\n", word.String(dawg.corpus))
	}
	prefixState := dawg.FindPrefix(word)
	prefixNode := prefixState.LastNode()
	if prefixNode.HasVertices() {
		dawg.ReplaceOrRegister(prefixNode)
	}
	suffix := word[prefixState.WordLength():]
	if len(suffix) > 0 {
		if prefixNode == dawg.finalNode {
			prefixNode = dawg.NewNode()
			prefixState.LastVertex().destination = prefixNode
		}
	}
	dawg.AddSuffix(prefixNode, suffix)
	return nil
}

func (dawg *Dawg) AddSuffix(node *Node, suffix Word) {
	if len(suffix) == 0 {
		return
	}
	if DAWG_TRACE {
		fmt.Printf("AddSuffix node#%d \"%s\"", node.id, suffix)
	}
	letter := suffix[0]
	suffix = suffix[1:]
	suffixNode := dawg.finalNode
	if len(suffix) > 0 {
		suffixNode = dawg.NewNode()
		dawg.AddSuffix(suffixNode, suffix)
	}
	dawg.AddVertex(node, letter, suffixNode, suffixNode == dawg.finalNode)
}

func (dawg *Dawg) fprintfNode(f io.Writer, node *Node) {
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
		fmt.Fprintf(f, "  +-- [%v] vertex#%v  letter:'%c' final:%v destination:%s \n", i, v.id, dawg.corpus.letterRune[v.letter], v.final, dest)
	}
}

func (dawg *Dawg) printState(state DawgState, args ...string) {
	dawg.fprintState(os.Stdout, state, args...)
}

func (dawg *Dawg) fprintState(f io.Writer, state DawgState, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	startNode := "node#nil"
	lastNode := "node#nil"
	if state.startNode != nil {
		startNode = fmt.Sprintf("node#%v", state.startNode.id)
	}
	if state.LastNode() != nil {
		lastNode = fmt.Sprintf("node#%v", state.LastNode().id)
	}

	fmt.Fprintf(f, "%sstate startNode:%s  word:\"%s\"  lastNode:%s\n", indent, startNode, state.Word().String(dawg.corpus), lastNode)
	for i, v := range state.vertices {
		dest := "!! nil destination !!"
		if v.destination != nil {
			dest = fmt.Sprintf("node#%v %s", v.destination.id, v.destination.vertexLetters.String(dawg.corpus))
		}
		fmt.Fprintf(f, "%s  +-- [%v] vertex#%v  letter:'%c' final:%v destination:%s \n", indent, i, v.id, dawg.corpus.letterRune[v.letter], v.final, dest)
	}
}

func (dawg *Dawg) printNode(node *Node) {
	dawg.fprintfNode(os.Stdout, node)
}

func (dawg *Dawg) print() {
	dawg.fprint(os.Stdout)
}

func (dawg *Dawg) fprint(f io.Writer) {
	fmt.Fprintf(f, "\n\n======== D A W G ========\n\n")
	dawg.fprintfRecurse(f, make(map[*Node]bool), dawg.initialState.startNode)
}

func (dawg *Dawg) fprintfRecurse(f io.Writer, printedNodes map[*Node]bool, node *Node) {
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

func (dawg *Dawg) printDot(seqno int, label string) {
	dotFileName := fmt.Sprintf("%s/%d_%s.gv", dotDir, seqno, label)
	dotFile, err := os.OpenFile(dotFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("error writing dot file : %s\n%v", dotFileName, err.Error())
		return
	}
	dawg.printDotSubtree(dotFile, label, dawg.initialState.startNode)
	dotFile.Close()
	fmt.Printf("wrote dot file : %s\n", dotFileName)
}

func (dawg *Dawg) printDotSubtree(f io.Writer, label string, node *Node) {
	fmt.Fprint(f, "digraph {\n")
	dawg.printfDotRecurse(f, make(map[*Node]bool), false, node)
	dawg.printfDotRecurse(f, make(map[*Node]bool), true, node)
	fmt.Fprintf(f, "label=\"\\n\\n%s\"\n", label)
	fmt.Fprint(f, "scale=\"0.5\"\n")
	fmt.Fprint(f, "}\n")

}

func (dawg *Dawg) printfDotRecurse(f io.Writer, printedNodes map[*Node]bool, printVertices bool, node *Node) {
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
					fmt.Fprintf(f, "%d -> %d [label=\" %c\" arrowhead=\"diamond\"]\n", node.id, v.destination.id, dawg.corpus.letterRune[v.letter])
				} else {
					fmt.Fprintf(f, "%d -> %d [label=\" %c\"]\n", node.id, v.destination.id, dawg.corpus.letterRune[v.letter])
				}
			}
			dawg.printfDotRecurse(f, printedNodes, printVertices, v.destination)
		}
	}
}
