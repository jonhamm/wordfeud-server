package dawg

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	. "wordfeud/corpus"
)

type node struct {
	id            nodeID
	registered    bool
	_crc          crc
	vertices      vertices
	vertexLetters LetterSet // if there is a vertex v in vertices then (vertexLetters & (1 << v.letter))!=0
}

type nodes []*node

func (node *node) crc() crc {
	if node._crc == 0 {
		var e error
		cs := crc32.NewIEEE()
		e = binary.Write(cs, binary.LittleEndian, node.vertexLetters)
		for _, v := range node.vertices {
			if e != nil {
				break
			}
			e = binary.Write(cs, binary.LittleEndian, v.final)
			if e == nil {
				e = binary.Write(cs, binary.LittleEndian, v.destination.crc())
			}
		}
		if e != nil {
			panic("node failure in ")
		}
		node._crc = crc(cs.Sum32())
	}
	return node._crc
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

func (node *node) findVertex(l Letter) (byte, *vertex) {
	if !node.vertexLetters.Test(l) {
		return byte(len(node.vertices)), nil
	}
	for i, v := range node.vertices {
		if v.letter == l {
			return byte(i), v
		}
	}
	panic("node inconsistent vertexLetters and vertices (node.findVertex)")
}

func (node *node) hasVertices() bool {
	if (len(node.vertices) == 0) != (node.vertexLetters == 0) {
		panic(fmt.Sprintf("node:%v inconsistent vertexLetters and vertices (node.hasVertices)", node.id))
	}
	return node.vertexLetters != 0
}

func (node *node) isSameState(otherNode *node) bool {
	if node == otherNode {
		return true
	}
	if node.vertexLetters != otherNode.vertexLetters {
		return false
	}
	if len(node.vertices) != len(otherNode.vertices) {
		panic(fmt.Sprintf("node:%v node:%v inconsistent vertexLetters and vertices (node.isSameState)", node.id, otherNode.id))
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

func (node *node) lastVertex() *vertex {
	if (len(node.vertices) == 0) != (node.vertexLetters == 0) {
		panic(fmt.Sprintf("node:%v inconsistent vertexLetters and vertices (node.LastVertexNode)", node.id))
	}
	if node.vertexLetters == 0 {
		return nil
	}
	return node.vertices[len(node.vertices)-1]
}
