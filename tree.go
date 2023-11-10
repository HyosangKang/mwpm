package mwpm

import (
	"gonum.org/v1/gonum/graph"
)

type Tree struct {
	g     graph.Weighted
	roots map[*Node]struct{}
	nodes map[int64]*Node
	temp  map[*Node]int64
	tight map[*Node]*Node // blossom -> node
}

func NewTree(wg graph.Weighted) *Tree {
	t := &Tree{
		g:     wg,
		roots: make(map[*Node]struct{}),
		nodes: make(map[int64]*Node),
		tight: make(map[*Node]*Node),
	}
	nodes := wg.Nodes()
	for nodes.Next() {
		nid := nodes.Node().ID()
		n := NewNode()
		n.temp = nid
		t.nodes[nid] = n
	}
	return t
}

func (t *Tree) Blossoms() []*Node {
	var nodes []*Node
	unique := make(map[*Node]struct{})
	for n := range t.roots {
		for _, m := range n.descendents() {
			b := m.Blossom()
			if _, ok := unique[b]; !ok {
				nodes = append(nodes, b)
				unique[b] = struct{}{}
			}
		}
	}
	return nodes
}

// set the blossom (or nodes) as a free node
func (t *Tree) SetFree(b *Node) {
	b.label = 0
	b.children = []*Node{}
	b.parent = nil
	for _, c := range b.cycle {
		c[0].BlossomWithin(b).label = 0
	}
}

// set the node n as tight within the blossom b
func (t *Tree) MakeTight(n, b *Node) {
	/* reorder the cycle to start from l */
	nb := n.BlossomWithin(b)
	for i, u := range nb.chain {
		if u == nb {
			nb.chain = append(nb.chain[i:], nb.chain[:i]...)
			nb.cycle = append(nb.cycle[i:], nb.cycle[:i]...)
			break
		}
	}
	for i := 1; i < len(nb.cycle); i += 2 {
		u, v := nb.cycle[i][0], nb.cycle[i][1]
		t.tight[u], t.tight[v] = v, u
		t.MakeTight(u, nb.chain[i])
		t.MakeTight(v, nb.chain[i+1])
	}
}

func (t *Tree) RemoveTight(s [2]*Node) {
	for _, u := range s {
		delete(t.tight, u)
		for u.blossom != nil {
			for _, c := range u.blossom.cycle {
				t.RemoveTight(c)
			}
		}
	}
}
