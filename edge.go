package mwpm

import "gonum.org/v1/gonum/graph"

var _ graph.WeightedEdge = Wedge{}

type Wedge struct {
	nodes  [2]Node
	weight float64
}

func (e Wedge) From() graph.Node {
	return e.nodes[0]
}

func (e Wedge) To() graph.Node {
	return e.nodes[1]
}

func (e Wedge) ReversedEdge() graph.Edge {
	return Wedge{nodes: [2]Node{e.nodes[1], e.nodes[0]}, weight: e.weight}
}

func (e Wedge) Weight() float64 {
	return e.weight
}
