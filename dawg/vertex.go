package dawg

import (
	. "wordfeud/corpus"
)

type vertex struct {
	id          vertexID
	letter      Letter
	final       bool
	destination *node
}

type vertices []*vertex
