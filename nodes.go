package mwpm

import "gonum.org/v1/gonum/graph"

var _ (graph.Nodes) = (*Nodes)(nil)

type Nodes struct {
	pos int
	lst []Node
}

func (ns *Nodes) Len() int {
	return len(ns.lst)
}

func (ns *Nodes) Next() bool {
	ns.pos++
	return ns.pos < len(ns.lst)
}

func (ns *Nodes) Node() graph.Node {
	return ns.lst[ns.pos]
}

func (ns *Nodes) Reset() {
	ns.pos = -1
}
