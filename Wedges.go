package mwpm

import "gonum.org/v1/gonum/graph"

var _ (graph.Edges) = (*Wedges)(nil)

type Wedges struct {
	pos int
	lst []Wedge
}

func (es *Wedges) Len() int {
	return len(es.lst)
}

func (es *Wedges) Next() bool {
	es.pos++
	return es.pos < len(es.lst)
}

func (es *Wedges) Edge() graph.Edge {
	return es.lst[es.pos]
}

func (es *Wedges) Reset() {
	es.pos = -1
}
