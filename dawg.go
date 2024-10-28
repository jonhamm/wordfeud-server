package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

const TRACE = true

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

type State struct {
	startNode *Node
	vertices  Vertices
	word      Word
}

var NullState = State{}

type Registry map[CRC]*Node

type Dawg struct {
	corpus       *Corpus
	rootNode     *Node
	finalNode    *Node
	initialState State
	registry     Registry
	nextNodeId   ID
	nextVertexId ID
}

func (dawg *Dawg) trace(format string, a ...any) {
	if TRACE {
		fmt.Fprintf(os.Stdout, format, a...)
	}
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
			e = binary.Write(cs, binary.LittleEndian, v.destination.CRC())
		}
		if e != nil {
			panic("Node failure in ")
		}
		node.crc = CRC(cs.Sum32())
	}
	return node.crc
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

func (state *State) LastVertex() *Vertex {
	if state.startNode == nil {
		return nil
	}
	n := len(state.vertices)
	if n == 0 {
		return nil
	}
	return state.vertices[n-1]
}

func (state *State) LastNode() *Node {
	lastVertex := state.LastVertex()
	if lastVertex == nil {
		return state.startNode
	}

	return lastVertex.destination
}

func MakeDawg(corpus *Corpus) (*Dawg, error) {

	dawg := &Dawg{
		corpus:       corpus,
		registry:     make(Registry),
		nextNodeId:   1,
		nextVertexId: 1,
	}

	dawg.rootNode = dawg.MakeNode()
	dawg.finalNode = dawg.MakeNode()
	dawg.Register(dawg.finalNode)
	dawg.initialState = State{startNode: dawg.rootNode, vertices: Vertices{}, word: Word{}}
	err := dawg.AddCorpus(corpus)
	if err != nil {
		return nil, err
	}
	return dawg, nil
}

func (dawg *Dawg) MakeNode() *Node {
	node := &Node{
		id:            dawg.nextNodeId,
		registered:    false,
		vertexLetters: 0,
		vertices:      Vertices{},
	}
	dawg.nextNodeId++
	return node
}

func (dawg *Dawg) MakeVertex(letter Letter, destination *Node, final bool) *Vertex {
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
	if TRACE {
		fmt.Printf("AddVertex node#%v letter:'%c' destination:node%v final:%v\n", node.id, dawg.corpus.letterRune[letter], destination.id, final)
		dawg.printNode(node)
	}
	if node.vertexLetters.test(letter) {
		panic(fmt.Sprintf("Node:%v trying to add vertex with an allready present letter ('%c') (Node.AddVertex)", node.id, dawg.corpus.letterRune[letter]))
	}
	if node.registered {
		panic(fmt.Sprintf("Node:%v trying to add vertex with letter ('%c') to registered node (Node.AddVertex)", node.id, dawg.corpus.letterRune[letter]))
	}
	vertex := dawg.MakeVertex(letter, destination, final)
	node.vertices = append(node.vertices, vertex)
	node.vertexLetters.set(vertex.letter)
	node.crc = 0 // invalidate crc
	if TRACE {
		dawg.printNode(node)
	}
	return vertex
}

func (dawg *Dawg) Lookup(node *Node) *Node {
	return dawg.registry[node.CRC()]
}

func (dawg *Dawg) Register(node *Node) {
	if TRACE {
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
	dawg.registry[node.CRC()] = node
	node.registered = true
}

func (dawg *Dawg) Transition(state State, letter Letter) State {
	if TRACE {
		fmt.Printf("Transition '%c' on state: \n", dawg.corpus.letterRune[letter])
		dawg.printState(state)
	}

	if state.startNode == nil {
		if TRACE {
			fmt.Printf("Transition '%c' on null state => NullState\n", dawg.corpus.letterRune[letter])
		}
		return NullState
	}
	node := state.LastNode()
	if node == nil {
		if TRACE {
			fmt.Printf("Transition '%c' on nil destination => NullState\n", dawg.corpus.letterRune[letter])
		}
		return NullState
	}

	_, v := node.FindVertex(letter)
	if v == nil {
		if TRACE {
			fmt.Printf("vertext for letter '%c' not found in node#%v  => NullState\n", dawg.corpus.letterRune[letter], node.id)
			dawg.printNode(node)
		}
		return NullState
	}

	transitionState := State{startNode: state.startNode, vertices: append(state.vertices, v), word: append(state.word, letter)}

	if TRACE {
		fmt.Printf("Transition '%c' in node#%v  => vertex#%v node#%v final:%v word:\"%s\"\n",
			dawg.corpus.letterRune[letter], node.id, v.id, v.destination.id, v.final, transitionState.word.String(dawg.corpus))
		dawg.printState(transitionState)
	}

	return transitionState
}

func (dawg *Dawg) Transitions(state State, word Word) State {
	if len(word) == 0 {
		return state
	}
	state = dawg.Transition(state, word[0])
	return dawg.Transitions(state, word[1:])
}

func (dawg *Dawg) CommonPrefix(word Word) State {
	state := dawg.initialState

	for _, l := range word {
		nextState := dawg.Transition(state, l)
		if nextState.startNode == nil {
			break
		}
		state = nextState
	}
	if TRACE {
		fmt.Printf("CommonPrefix \"%s\" => \n", word.String(dawg.corpus))
		dawg.printState(state)
	}
	return state
}

func (dawg *Dawg) ReplaceOrRegister(node *Node) {
	if TRACE {
		fmt.Printf("ReplaceOrRegister node#%v\n", node.id)
		dawg.printNode(node)
	}

	lastVertex := node.LastVertex()

	if TRACE {
		hasVertices := "has no vertices"
		if lastVertex.destination.HasVertices() {
			hasVertices = "has vertices"
		}
		fmt.Printf("lastvertex#%v('%c').destination: node#%v %s\n", lastVertex.id, dawg.corpus.letterRune[lastVertex.letter], node.id, hasVertices)
	}

	if lastVertex.destination.HasVertices() {
		dawg.ReplaceOrRegister(lastVertex.destination)
	}

	if TRACE {
		fmt.Printf("-->ReplaceOrRegister node#%v\n", node.id)
	}

	registryNode := dawg.Lookup(lastVertex.destination)
	if registryNode != nil {
		lastVertex.destination = registryNode

		if TRACE {
			fmt.Printf("lastvertex#%v('%c').destination <= registry node#%v\n", lastVertex.id, dawg.corpus.letterRune[lastVertex.letter], registryNode.id)
		}

	} else {
		dawg.Register(lastVertex.destination)
	}
}

func (dawg *Dawg) AddCorpus(corpus *Corpus) error {
	for _, w := range corpus.words {
		err := dawg.AddWord((w))
		if err != nil {
			return err
		}
	}
	dawg.ReplaceOrRegister(dawg.rootNode)
	if TRACE {
		dawg.print()
	}
	return nil
}

func (dawg *Dawg) AddWord(word Word) error {
	if TRACE {
		dawg.trace("\n\nAddWord \"%s\"\n\n", word.String(dawg.corpus))
	}
	prefixState := dawg.CommonPrefix(word)
	prefixNode := prefixState.LastNode()
	if prefixNode.HasVertices() {
		dawg.ReplaceOrRegister(prefixNode)
	}
	suffix := word[len(prefixState.word):]
	if len(suffix) > 0 {
		if prefixNode == dawg.finalNode {
			prefixNode = dawg.MakeNode()
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
	letter := suffix[0]
	suffix = suffix[1:]
	suffixNode := dawg.finalNode
	if len(suffix) > 0 {
		suffixNode = dawg.MakeNode()
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
		dest := "!! nil destination !!"
		if v.destination != nil {
			dest = fmt.Sprintf("node#%v", v.destination.id)
		}
		fmt.Fprintf(f, "  +-- [%v] vertex#%v  letter:'%c' final:%v destination:%s \n", i, v.id, dawg.corpus.letterRune[v.letter], v.final, dest)
	}
}

func (dawg *Dawg) printState(state State) {
	dawg.fprintState(os.Stdout, state)
}

func (dawg *Dawg) fprintState(f io.Writer, state State) {
	startNode := "node#nil"
	if state.startNode != nil {
		startNode = fmt.Sprintf("node#%v\n", state.startNode.id)
	}
	fmt.Fprintf(f, "state startNode:%s\n", startNode)
	for i, v := range state.vertices {
		dest := "!! nil destination !!"
		if v.destination != nil {
			dest = fmt.Sprintf("node#%v", v.destination.id)
		}
		fmt.Fprintf(f, "  +-- [%v] vertex#%v  letter:'%c' final:%v destination:%s \n", i, v.id, dawg.corpus.letterRune[v.letter], v.final, dest)
	}
}

func (dawg *Dawg) printNode(node *Node) {
	dawg.fprintfNode(os.Stdout, node)
}

func (dawg *Dawg) print() {
	dawg.fprint(os.Stdout)
}

func (dawg *Dawg) fprint(f io.Writer) {
	nodes := make(map[*Node]bool)
	fmt.Fprintf(f, "\n\n======== D A W G ========\n\n")
	dawg.fprintfRecurse(f, nodes, dawg.initialState.startNode)
}

func (dawg *Dawg) fprintfRecurse(f io.Writer, nodes map[*Node]bool, node *Node) {
	if nodes[node] {
		return
	}
	dawg.fprintfNode(f, node)
	nodes[node] = true
	for _, v := range node.vertices {
		if v.destination != nil {
			dawg.fprintfRecurse(f, nodes, v.destination)
		}
	}
}
