package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

type CRC uint32
type ID uint32

type Vertex struct {
	id          ID
	letter      Letter
	finalNode   bool
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
	node  *Node
	final bool
	word  Word
}

var NullState = State{nil, false, Word{}}

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
	if node.vertexLetters.test(l) {
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

func (node *Node) AddVertex(vertex *Vertex) *Vertex {
	if node.vertexLetters.test(vertex.letter) {
		panic(fmt.Sprintf("Node:%v trying to add vertex with an allready present letter ('%v') (Node.AddVertex)", node.id, vertex.letter))
	}
	if node.registered {
		panic(fmt.Sprintf("Node:%v trying to add vertex with letter ('%v') to registered node (Node.AddVertex)", node.id, vertex.letter))
	}
	node.vertices = append(node.vertices, vertex)
	node.vertexLetters.set(vertex.letter)
	node.crc = 0 // invalidate crc
	return vertex
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
	dawg.initialState = State{node: dawg.rootNode, final: false, word: Word{}}
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

func (dawg *Dawg) MakeVertex(letter Letter, destination *Node, finalNode bool) *Vertex {
	vertex := &Vertex{
		id:          dawg.nextVertexId,
		letter:      letter,
		destination: destination,
		finalNode:   finalNode,
	}
	dawg.nextVertexId++
	return vertex
}

func (dawg *Dawg) Lookup(node *Node) *Node {
	return dawg.registry[node.CRC()]
}

func (dawg *Dawg) Register(node *Node) {
	n := dawg.Lookup(node)
	if n != nil {
		panic(fmt.Sprintf("Dawg trying to register node:%v allready registered as node:%v (Dawg.Register): ", node.id, n.id))
	}
	dawg.registry[node.CRC()] = node
}

func (dawg *Dawg) Transition(state State, letter Letter) State {
	if state.node == nil {
		return NullState
	}
	_, v := state.node.FindVertex(letter)
	if v == nil {
		return NullState
	}
	return State{node: v.destination, final: v.finalNode, word: append(state.word, letter)}
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
		if nextState.node == nil {
			break
		}
		state = nextState
	}

	return state
}

func (dawg *Dawg) ReplaceOrRegister(node *Node) {
	lastVertex := node.LastVertex()
	if lastVertex.destination.HasVertices() {
		dawg.ReplaceOrRegister(lastVertex.destination)
	}
	registryNode := dawg.Lookup(lastVertex.destination)
	if registryNode != nil {
		lastVertex.destination = registryNode
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
	return nil
}

func (dawg *Dawg) AddWord(word Word) error {
	prefixState := dawg.CommonPrefix(word)
	if prefixState.node.HasVertices() {
		dawg.ReplaceOrRegister(prefixState.node)
	}
	suffix := word[len(prefixState.word):]
	dawg.AddSuffix(prefixState.node, suffix)
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
	node.AddVertex(dawg.MakeVertex(letter, suffixNode, suffixNode == dawg.finalNode))
}

func (dawg *Dawg) AddVertex(node *Node, letter Letter, destination *Node, finalNode bool) *Vertex {
	return node.AddVertex(dawg.MakeVertex(letter, destination, finalNode))
}
