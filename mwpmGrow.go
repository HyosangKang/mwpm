package mwpm

import (
	"fmt"
)

// Grow takes an weighted edge e to attach a blossom edge to the blossom graph.
// The opration is depicted as
//	o (+) u                o (+) u
//	               --->     \ e
// (0) o = o (0)         (-) o = o (+)
//	v   w                 v   w
// Grow makes the tree of u to 'grow'.
// No new match is made in Grow.
// The fields that needs updates are:
// 0. labels of v, w
// 1. edges bewteen u, v
// 2. parent, root, children of u, v, w

func (g *Graph) Grow(e Wedge) {
	u := g.Blossom(e.From().ID())
	v := g.Blossom(e.To().ID())
	if g.label[u] == 0 {
		u = g.Blossom(e.To().ID())
		v = g.Blossom(e.From().ID())
	}
	w := g.MatchTo(v)
	if w == -1 {
		fmt.Printf("No blossom is matched to [%d]\n", v)
		return
	}

	// Update the labels of v, w.
	g.label[v] = -1
	g.label[w] = 1

	// Create an edge between u, v.
	g.SetEdge(u, v, e)

	// Updates tree.
	g.root[v] = g.root[u]
	g.root[w] = g.root[u]
	g.children[u] = append(g.children[u], v)
	g.children[v] = []int64{w}
	g.parent[v] = u
	g.parent[w] = v
}

// MatchTo finds a matched blossom to the given blossom.
func (g *Graph) MatchTo(u int64) int64 {
	if fn, ok := g.edges[u]; ok {
		for n, _ := range fn {
			return n
		}
	}
	return -1
}
