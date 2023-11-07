package mwpm

import "gonum.org/v1/gonum/graph"

var _ (graph.Edges) = (*Edges)(nil)

type Edges struct {
	pos int
	lst []Wedge
}

func (es *Edges) Len() int {
	return len(es.lst)
}

func (es *Edges) Next() bool {
	es.pos++
	return es.pos < len(es.lst)
}

func (es *Edges) Edge() graph.Edge {
	return es.lst[es.pos]
}

func (es *Edges) Reset() {
	es.pos = -1
}
