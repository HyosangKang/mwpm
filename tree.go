package mwpm

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

// Tree consists of
// 0. the weighted graph wg: only for referece. No change is made to wg whatsoever.
// 1. root(int64): the root of the tree. The root of a root is itself.
// 2. parent(int64): the parent of the blossom. The parent of a root is itself.
// 3. children([]int64): the children of the blossom. It is a slice of blossoms.
// 4. label(int): one of +1, 0, -1.
// 5. dval(float64): the dual value of blossom
// 6. edges(BlossomEdge): each edge between blossoms is made by an edge in wg.
// 7. blossom(int64): the blossom containing sub-blossom. -1 if not contained in any sup-blossom.
// 8. cycle([]int64): the slice of sub-blossoms. The cyclic order should be preserved.
// 9. nodes([]int64): the slice of all nodes (not blossoms) in the blossom.
type Tree struct {
	roots map[*Node]struct{}
	nodes map[int64]*Node
	edges map[int64]map[int64]float64
	match map[*Node]*Node
}

// NewBlossomGraphFrom takes a weighted undirected graph as input
// and retuns a blossom graph with initialization.
// The blossom graph contains only nodes, and no blossom edge.
// Each node becomes a blossom of itself.
// The labels are set to +1.
// The root, parent of each blossom is itself and empty children.
func NewTree(wg graph.Weighted) *Tree {
	t := &Tree{
		roots: make(map[*Node]struct{}),
		nodes: make(map[int64]*Node),
		edges: make(map[int64]map[int64]float64),
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

func (t *Tree) Nodes() []*Node {
	var nodes []*Node
	for n := range t.roots {
		nodes = append(nodes, n.Descendants()...)
	}
	return nodes
}

func (t *Tree) Free(n *Node) {
	if n.label > 0 {
		delete(t.roots, n)
	}
	n.label = 0
}
