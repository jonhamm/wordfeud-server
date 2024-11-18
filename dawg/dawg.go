package dawg

import (
	"fmt"
	"io"
	"os"
	. "wordfeud/context"
	. "wordfeud/corpus"
)

var DAWG_TRACE = false

const dotDir = "/tmp/scrabble_dot/"

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
}

func NewDawg(content CorpusContent, options Options) (Dawg, error) {
	corpus := content.Corpus()

	dawg := &_Dawg{
		_options: options,
		corpus:   corpus,
	}

	builder := dawgBuilder{
		dawg:         dawg,
		registry:     make(Registry),
		nextNodeId:   1,
		nextVertexId: 1,
	}
	dawg._rootNode = builder.newNode()
	dawg._finalNode = builder.newNode()
	dawg.nullState = &_DawgState{dawg: dawg, startNode: nil, vertices: vertices{}}
	dawg._initialState = &_DawgState{dawg: dawg, startNode: dawg._rootNode, vertices: vertices{}}
	builder.register(dawg._finalNode)

	err := builder.addCorpus(content)
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

func (dawg *_Dawg) fprintfNode(f io.Writer, node *node) {
	if node == nil {
		fmt.Fprint(f, "node#nil\n")
		return
	}
	fmt.Fprintf(f, "node#%v  crc:%v registered:%v vertexLetters:%s\n", node.id, node._crc, node.registered, node.vertexLetters.String(dawg.corpus))
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
