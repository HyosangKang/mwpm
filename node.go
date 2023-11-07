package mwpm

import "gonum.org/v1/gonum/graph"

var _ (graph.Node) = Node{}

const Eps = 1e-6

type Node struct {
	label            int8
	id, root, parent int64
	cycle, nodes     []int64
	dval             float64
}

func (n Node) ID() int64 {
	return n.id
}

func NewNode() Node {
	return Node{
		id:     -1,
		root:   -1,
		parent: -1,
	}
}

func (n Node) DualVal() float64 {
	return n.dval
}

func (n Node) Label() int8 {
	return n.label
}

func (n Node) Root() int64 {
	return n.root
}

func (n Node) IsBlossom() bool {
	return len(n.cycle) > 0
}
