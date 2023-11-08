package mwpm

import (
	"gonum.org/v1/gonum/graph"
)

type Tree struct {
	roots map[*Node]struct{}
	nodes map[int64]*Node
	edges map[int64]map[int64]float64
	tight map[*Node]*Node // blossom -> node
}

func NewTree(wg graph.Weighted) *Tree {
	t := &Tree{
		roots: make(map[*Node]struct{}),
		nodes: make(map[int64]*Node),
		edges: make(map[int64]map[int64]float64),
		tight: make(map[*Node]*Node),
	}
	nodes := wg.Nodes()
	for nodes.Next() {
		nid := nodes.Node().ID()
		n := NewNode()
		t.roots[n] = struct{}{}
		t.nodes[nid] = n
		modes := wg.Nodes()
		for modes.Next() {
			mid := modes.Node().ID()
			if wg.HasEdgeBetween(nid, mid) {
				if _, ok := t.edges[nid]; !ok {
					t.edges[nid] = make(map[int64]float64)
				}
				t.edges[nid][mid], _ = wg.Weight(nid, mid)
			}
		}
	}
	return t
}

func (t *Tree) Blossoms() map[*Node]struct{} {
	var nodes []*Node
	for n := range t.roots {
		nodes = append(nodes, n.descendants()...)
	}
	unique := make(map[*Node]struct{})
	for _, n := range nodes {
		b := n.Blossom()
		if _, ok := unique[b]; !ok {
			unique[b] = struct{}{}
		}
	}
	return unique
}

// turn the node into a free node.
func (t *Tree) Free(n *Node) {
	if n.label > 0 {
		delete(t.roots, n)
	}
	n.label = 0
}
