package mwpm

import "gonum.org/v1/gonum/graph"

var _ graph.Edge = Bedge{}

type Bedge struct {
	nodes [2]Node
}

func (e Bedge) From() graph.Node {
	return e.nodes[0]
}

func (e Bedge) To() graph.Node {
	return e.nodes[1]
}

func (e Bedge) ReversedEdge() graph.Edge {
	return Bedge{nodes: [2]Node{e.nodes[1], e.nodes[0]}}
}
