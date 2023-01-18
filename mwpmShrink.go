package mwpm

import (
	"gonum.org/v1/gonum/graph"
)

// Shrink makes a new blossom consists of nodes in a tree,
// where two (+) sub-blossoms are connected by edge e.
//     (-) o p
//         |                     o p              o (+)
//     (+) o                     |              /   \
//       /   \        ---->      o b       (-) o  b  o (-)
//  (-) o     o (-)             / \            |     |
//      |  e  |            (-) o   o (-)       o  -  o
//  o - o  - o - o                            (+)   (+)
// (-) (+)  (+) (-)
// It does not remove the nodes, but there are changes in tree and blossom edges.
func (g *BlossomGraph) Shrink(e graph.WeightedEdge) {

	// u, v are blossoms containing nodes of e.
	// (Both blossoms are (+) labeled and lie in the same tree.)
	u := g.Blossom(e.From().ID())
	v := g.Blossom(e.To().ID())

	// Find the common anscester of u and v.
	// Then create a cycle of blossoms in the followwing order: u -> ... -> p -> ... -> v
	// Also create edge between u and v
	p, cycle := g.Cycle(u, v)
	g.SetEdge(u, v, e)

	// Create new blossom b consists of blossoms in the cycle.
	// The label of b is set to +1, and dual value 0.
	// The field blossom, nodes are initialized set as well.
	b := int64(len(g.label))
	g.label[b] = +1
	g.cycle[b] = cycle
	g.nodes[b] = []int64{}
	for _, n := range cycle {
		g.nodes[b] = append(g.nodes[b], g.nodes[n]...)
	}
	g.dval[b] = 0

	// The blossoms in the cycle now belong to the blossom b.
	// All labels are set to 0.
	for _, n := range cycle {
		g.blossom[n] = b
		g.LabelAsZero(n)
	}

	// If p is not the root of the tree, remove the edges between parent of p (pp) and p,
	// and create a edge between b and pp.
	// Remove all edges between blossoms in the cycle and their children,
	// and replace by edges between b and the children
	pp := g.parent[p]
	if pp != p {
		be := g.edges[pp][p]
		delete(g.edges[pp], p)
		delete(g.edges[p], pp)
		g.SetEdge(pp, b, be.e)
	}
	for _, n := range cycle {
		for _, m := range g.children[n] {
			if ContainsInt64(cycle, m) {
				continue
			}
			be := g.edges[n][m]
			delete(g.edges[n], m)
			delete(g.edges[m], n)
			g.SetEdge(b, m, be.e)
		}
	}

	// Set the heritage of the new blossom.
	// (Change root, parent, children info.)
	if pp != p {
		g.root[b] = g.root[p]
		g.parent[b] = pp
		g.children[pp] = []int64{b}
	} else {
		g.parent[b] = b
		g.root[b] = b
	}
	for _, n := range cycle {
		for _, m := range g.children[n] {
			if ContainsInt64(cycle, m) {
				continue
			}
			if _, ok := g.children[b]; ok {
				g.children[b] = append(g.children[b], m)
			} else {
				g.children[b] = []int64{m}
			}
			g.parent[m] = b
		}
	}
	if g.root[b] == b {
		g.ChangeRootFrom(b, b)
	}

	// Also, remove the tree structure of nodes in the cycle.
	for _, n := range cycle {
		delete(g.root, n)
		delete(g.parent, n)
		delete(g.children, n)
	}

	// Remove all matches in b.
	g.UnMatchBlossom(b)

	// Set new edges of b and match to its parent (if exists.)
	if g.root[b] != b {
		g.MatchEdgeBetween(b, g.parent[b])
	}
}

// CommonParent finds the common ansester of u, v and cycle of blossoms from u to p and to v.
// For example it returns (2, [0,1,2,4,3]) if the tree looks like:
//        o  [2]
//      /   \
// [1] o     o [4]
//     |     |
// [0] o     o [3]
//     u     v
func (g *BlossomGraph) Cycle(u, v int64) (int64, []int64) {
	h1 := g.Heritage(u)
	h2 := g.Heritage(v)

	var i1, i2 int
	found := false
	for i1 = 0; i1 < len(h1); i1++ {
		for i2 = 0; i2 < len(h2); i2++ {
			if h1[i1] == h2[i2] {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return -1, []int64{}
	}
	rev := []int64{}
	for i := i2; i > -1; i-- {
		rev = append(rev, h2[i])
	}
	return h1[i1], append(h1[:i1], rev...)
}

func (g *BlossomGraph) ChangeRootFrom(b, r int64) {
	g.root[b] = r
	if _, ok := g.children[b]; !ok {
		return
	}
	for _, n := range g.children[b] {
		g.ChangeRootFrom(n, r)
	}
}
