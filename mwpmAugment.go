package mwpm

import (
	"gonum.org/v1/gonum/graph"
)

// Augment increases the matching. It is desribed in the following pictorial way:
// (+) o        o (+)                o     o
//
//	   \     /   \                  I     I
//	(-) o   o (-) o (-)   ----->    o     o   o
//	    I e I     I                           I
//	(+) o - o (+) o (+)             o +-+ o   o
//	    u   v
//
// In summary, we operates three things:
// 0. Invert matching along the hierarchies from u and v
// 1. Create a 'matched' blossom edge between u and v
// 2. Remove trees of u, v. This also sets all labels to 0.
// When create a new match between two blossoms,
// matchings in each blossom occurs recursively.
// For example, if a new match is made to the node [0],
// then new matchs are created in the blossom of [0]
//
//	      o - o                    o +-+ o
//	[0] /       \                /         \
//
// +-+ o  -  o   o          +-+ o  -  o     o
//
//	|\   /    |    ---->     |\    I     I
//	|  o      o              |  -  o     o
//	 \       /                \         /
//	   o - o                    o +-+ o
//
// All matchings in two blossoms must be wiped before create a match between two.
func (g *Graph) Augment(e graph.WeightedEdge) {
	u := g.Blossom(e.From().ID())
	v := g.Blossom(e.To().ID())

	// Find anscesters of two blossoms and remove all matchings
	// and create a match beween blossoms in inverted way.
	g.UnMatchBlossom(u)
	g.UnMatchBlossom(v)
	h := g.Heritage(u)
	for i := 0; i < len(h)-1; i++ {
		g.UnMatchEdgeBetween(h[i], h[i+1])
	}
	for i := 1; i < len(h)-1; i += 2 {
		g.MatchEdgeBetween(h[i], h[i+1])
	}
	h = g.Heritage(v)
	for i := 0; i < len(h)-1; i++ {
		g.UnMatchEdgeBetween(h[i], h[i+1])
	}
	for i := 1; i < len(h)-1; i += 2 {
		g.MatchEdgeBetween(h[i], h[i+1])
	}

	// Create a matched blossom edge between u and v.
	g.SetEdge(u, v, e)
	g.MatchEdgeBetween(u, v)

	// Remove tree. If an edge in tree is not a matached edge, remove it too.
	// All labels are set to 0 as well.
	g.RemoveTree(g.root[u])
	g.RemoveTree(g.root[v])
}

// Heriatage returns all anscesters from u to its root in the blossom graph.
func (g *Graph) Heritage(u int64) []int64 {
	if g.parent[u] == u {
		return []int64{u}
	}
	return append([]int64{u}, g.Heritage(g.parent[u])...)
}

func (g *Graph) SetEdge(u, v int64, e graph.WeightedEdge) {
	be := Bedge{
		e:     e,
		match: false,
	}
	if _, ok := g.edges[u]; ok {
		g.edges[u][v] = be
	} else {
		g.edges[u] = map[int64]Bedge{v: be}
	}
	if _, ok := g.edges[v]; ok {
		g.edges[v][u] = be
	} else {
		g.edges[v] = map[int64]Bedge{u: be}
	}
}

// UnMatchBlossom recursively removes all matchings in blossom edges in the blossom.
func (g *Graph) UnMatchBlossom(u int64) {
	if len(g.cycle[u]) == 1 {
		return
	}
	cycle := g.cycle[u]
	for i := 0; i < len(cycle)-1; i++ {
		be := g.edges[cycle[i]][cycle[i+1]]
		be = Bedge{
			e:     be.e,
			match: false,
		}
		g.edges[cycle[i]][cycle[i+1]] = be
		g.edges[cycle[i+1]][cycle[i]] = be
		g.UnMatchBlossom(cycle[i])
	}
	last := int64(len(cycle)) - 1
	be := g.edges[cycle[last]][cycle[0]]
	be = Bedge{
		e:     be.e,
		match: false,
	}
	g.edges[cycle[last]][cycle[0]] = be
	g.edges[cycle[0]][cycle[last]] = be
	g.UnMatchBlossom(cycle[last])
}

// UnMatchEdgeBetween removes matching bewteen blossoms u and v,
// and wipes all matchs inside each blossom.
// It assumes a blossom edge between u, v, and does not remove the edge.
func (g *Graph) UnMatchEdgeBetween(u, v int64) {
	be := g.edges[u][v]
	be = Bedge{
		e:     be.e,
		match: false,
	}
	g.edges[u][v] = be
	g.edges[v][u] = be
	g.UnMatchBlossom(u)
	g.UnMatchBlossom(v)
}

// MatchEdgeBetween set the blossom edge between u, v,
// and recursively create matchs in each blossoms.
// There must be a blossom edges between u and v, a priori.
func (g *Graph) MatchEdgeBetween(u, v int64) {
	// fmt.Printf("Match edge between [%d] and [%d]\n", u, v)
	be := g.edges[u][v]
	be = Bedge{
		e:     be.e,
		match: true,
	}
	g.edges[u][v] = be
	g.edges[v][u] = be

	n := be.e.From().ID()
	m := be.e.To().ID()

	if ContainsInt64(g.nodes[u], n) {
		g.MatchBlossom(u, n)
		g.MatchBlossom(v, m)
	} else {
		g.MatchBlossom(u, m)
		g.MatchBlossom(v, n)
	}
}

// MatchBlossom matches the blossom edges in the blossom u provided that
// node n is matched to a node out side of the blossom u.
// It recursively create matches until no further match is available.
func (g *Graph) MatchBlossom(u, n int64) {
	// fmt.Printf("Match blossom [%d] from [%d]\n", u, n)
	if len(g.cycle[u]) == 1 {
		return
	}

	// Reorder the cycle of u that starts from the sub-blossom containing n.
	var i int
	cycle := g.cycle[u]
	for i = 0; i < len(cycle); i++ {
		if ContainsInt64(g.nodes[cycle[i]], n) {
			break
		}
	}
	cycle = append(cycle[i:], cycle[:i]...)

	// Match blossom which contains the original node.
	g.MatchBlossom(cycle[0], n)

	// fmt.Printf("Reordered cycle as: %+v\n", cycle)
	for i = 1; i < len(cycle)-1; i += 2 {
		g.MatchEdgeBetween(cycle[i], cycle[i+1])
	}
}

// Contains checks whether the slice ns ns contains n
func ContainsInt64(ns []int64, n int64) bool {
	for _, m := range ns {
		if m == n {
			return true
		}
	}
	return false
}

// RemoveTree removes all information about tree from the root r.
// It removes the edge between parent and child if they are connected by an unmatched edge.
// It sets all labels of blossoms to 0. (and all nodes in it, recursively.)
func (g *Graph) RemoveTree(r int64) {
	delete(g.root, r)
	delete(g.parent, r)
	g.LabelAsZero(r)
	if _, ok := g.children[r]; !ok {
		return
	}
	for _, n := range g.children[r] {
		be := g.edges[r][n]
		if be.match == false {
			delete(g.edges[r], n)
			delete(g.edges[n], r)
		}
		g.RemoveTree(n)
	}
	delete(g.children, r)
}

// LabelAsZero sets all labels of nodes/sub-blossoms in b as 0.
func (g *Graph) LabelAsZero(b int64) {
	g.label[b] = 0
	if len(g.cycle[b]) == 1 {
		return
	}
	for _, n := range g.cycle[b] {
		g.LabelAsZero(n)
	}
}
