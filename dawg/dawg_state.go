package dawg

import (
	"fmt"
	"io"
	"os"
	"slices"
	. "wordfeud/corpus"
)

type DawgState interface {
	Dawg() Dawg
	Valid() bool
	Final() bool
	Word() Word
	WordLength() int
	Transition(Letter) DawgState
	Transitions(Word) DawgState
	ValidContinuations(suffixes ...Word) LetterSet
	Print(args ...string)
	FprintState(f io.Writer, args ...string)
	lastNode() *node
	lastVertex() *vertex
}

type _DawgState struct {
	dawg      *_Dawg
	startNode *node
	vertices  vertices
}

func (state *_DawgState) Dawg() Dawg {
	return state.dawg
}

func (state *_DawgState) Valid() bool {
	return state != nil && state.startNode != nil
}

func (state *_DawgState) Final() bool {
	if !state.Valid() {
		return false
	}
	v := state.lastVertex()
	return v != nil && v.final
}

func (state *_DawgState) Transition(letter Letter) DawgState {
	dawg := state.dawg
	if !state.Valid() {
		return dawg.nullState
	}
	if DAWG_TRACE {
		fmt.Printf("Transition '%c' on state: \n", dawg.corpus.LetterToRune(letter))
		state.Print()
	}

	if state.startNode == nil {
		if DAWG_TRACE {
			fmt.Printf("Transition '%c' on null state => dawg.nullState\n", dawg.corpus.LetterToRune(letter))
		}
		return dawg.nullState
	}
	node := state.lastNode()
	if node == nil {
		if DAWG_TRACE {
			fmt.Printf("Transition '%c' on nil destination => dawg.nullState\n", dawg.corpus.LetterToRune(letter))
		}
		return dawg.nullState
	}

	_, v := node.findVertex(letter)
	if v == nil {
		if DAWG_TRACE {
			fmt.Printf("vertext for letter '%c' not found in node#%v  => dawg.nullState\n", dawg.corpus.LetterToRune(letter), node.id)
			dawg.printNode(node)
		}
		return dawg.nullState
	}

	transitionState := &_DawgState{
		dawg:      dawg,
		startNode: state.startNode,
		vertices:  slices.Concat(state.vertices, vertices{v}),
	}

	if DAWG_TRACE {
		fmt.Printf("Transition '%c' in node#%v  => vertex#%v node#%v final:%v word:\"%s\"\n",
			dawg.corpus.LetterToRune(letter), node.id, v.id, v.destination.id, v.final, transitionState.Word().String(dawg.corpus))
		transitionState.Print()
	}

	return transitionState
}

func (state *_DawgState) Transitions(word Word) DawgState {
	dawg := state.dawg
	if !state.Valid() {
		return dawg.nullState
	}
	if len(word) == 0 {
		return state
	}
	transitionState := state.Transition(word[0])
	transitionState = transitionState.Transitions(word[1:])
	return transitionState
}

func (state *_DawgState) ValidContinuations(suffixes ...Word) LetterSet {
	options := state.dawg.options()
	corpus := state.dawg.Corpus()
	validContinuations := NullLetterSet
	if !state.Valid() {
		return validContinuations
	}
	if len(suffixes) == 0 {
		suffixes = Words{Word{}}
	}
	endNode := state.lastNode()
	for _, suffixWord := range suffixes {
		if len(suffixWord) > 0 {
			if options.Debug > 0 {
				fmt.Printf("   suffix: %s\n", suffixWord.String(corpus))
			}
			for _, v := range endNode.vertices {
				s := &_DawgState{dawg: state.dawg, startNode: v.destination, vertices: vertices{}}
				suffix := s.Transitions(suffixWord)
				if suffix.Valid() {
					if suffix.Final() {
						validContinuations.Set(v.letter)
					}
				}
			}
		} else {
			for _, v := range endNode.vertices {
				if v.final {
					validContinuations.Set(v.letter)
				}
			}

		}
	}
	return validContinuations
}

func (state *_DawgState) Word() Word {
	word := make(Word, len(state.vertices))
	for i, v := range state.vertices {
		word[i] = v.letter
	}
	return word
}

func (state *_DawgState) WordLength() int {
	return len(state.vertices)
}

func (state *_DawgState) lastVertex() *vertex {
	if state.startNode == nil {
		return nil
	}
	n := len(state.vertices)
	if n == 0 {
		return nil
	}
	return state.vertices[n-1]
}

func (state *_DawgState) lastNode() *node {
	lastVertex := state.lastVertex()
	if lastVertex == nil {
		return state.startNode
	}

	return lastVertex.destination
}

func (state *_DawgState) Print(args ...string) {
	state.FprintState(os.Stdout, args...)
}

func (state *_DawgState) FprintState(f io.Writer, args ...string) {
	dawg := state.dawg
	corpus := dawg.Corpus()
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	startNode := "node#nil"
	lastNode := "node#nil"
	if state.startNode != nil {
		startNode = fmt.Sprintf("node#%v", state.startNode.id)
	}
	if state.lastNode() != nil {
		lastNode = fmt.Sprintf("node#%v", state.lastNode().id)
	}

	fmt.Fprintf(f, "%sstate startNode:%s  word:\"%s\"  lastNode:%s\n", indent, startNode, state.Word().String(dawg.corpus), lastNode)
	for i, v := range state.vertices {
		dest := "!! nil destination !!"
		if v.destination != nil {
			dest = fmt.Sprintf("node#%v %s", v.destination.id, v.destination.vertexLetters.String(dawg.corpus))
		}
		fmt.Fprintf(f, "%s  +-- [%v] vertex#%v  letter:'%c' final:%v destination:%s \n", indent, i, v.id, corpus.LetterToRune(v.letter), v.final, dest)
	}
}
