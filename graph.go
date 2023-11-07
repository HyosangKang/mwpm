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

// Match returns all pairs of matched nodes.
// It counts the same match twice, only for simplicity
func (g *Tree) Match() [][2]int64 {
	seen := make(map[[2]int64]struct{})
	match := [][2]int64{}
	for _, fn := range g.edges {
		for _, be := range fn {
			if be.match {
				n := be.e.From().ID()
				m := be.e.To().ID()
				if _, ok := seen[[2]int64{n, m}]; !ok {
					match = append(match, [2]int64{n, m})
					seen[[2]int64{n, m}] = struct{}{}
					seen[[2]int64{m, n}] = struct{}{}
				}
			}
		}
	}
	return match
}

// TightEdge returns a tight edge together with case number.
// An edge e is called tight if slack(e) = 0 (See Slack for its formula.)
// Since augment occurs when the case number is 1, it has priority among others.
func (t *Tree) TightEdge() (uid, vid int64) {
	for nid, edges := range t.edges {
		for mid, _ := range edges {
			if mid <= nid {
				continue
			}
			if t.Slack([2]int64{nid, mid}) == 0 {
				return nid, mid
			}
		}
	}
	return -1, -1
}

// NegBlossom returns a blossom (>3 nodes) with -1 label and 0 dual value.
// It returns -1 if no such blossom exists.
func (g *Tree) NegBlossom() int64 {
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node().(Node)
		if n.IsBlossom() && n.Label() == -1 && n.DualVal() < Eps {
			return n.ID()
		}
	}
	return -1
}

// Delta returns the minimum value among four types of values:
// slack(u,v) if the edge e=(u,v) is of case 0;
// slack(u,v)/2 if the edge e=(u,v) is of case 1 or 2;
// dualValue(b) if b is a blossom (>3 nodes) of -1 label.
func (g *Tree) Delta() float64 {
	var d float64
	d = math.Inf(+1)

	all := g.wg.WeightedEdges()
	for all.Next() {
		e := all.WeightedEdge()
		c := g.Case(e)
		if c != -1 {
			s := g.Slack(e)
			if c > 0 {
				s /= 2
			}
			if d > s {
				d = s
			}
		}
	}
	for n, l := range g.label {
		if l == -1 {
			if len(g.cycle[n]) > 1 {
				if d > g.dval[n] {
					d = g.dval[n]
				}
			}
		}
	}
	return d
}

// The slack of an edge e = (u, v) is
// slack(e) = weight(e) - dualValue(u) - dualValue(v)
// Here, u and v are nodes not blossoms.
func (g *Tree) Slack(ids [2]int64) float64 {
	s := g.edges[ids[0]][ids[1]]
	for i := 0; i < 2; i++ {
		s -= g.nodes[ids[i]].Blossom().dval
	}
	return s
}

// Case returns the case of an edge e = (u, v) as below
// Let n, m be the top blossom containing u, v respectively.
// If the labels of n, m are
// (+1, 0) or (0, +1), then the case number is 0;
// (+1, +1) and u, v lie in different tree, then the case number is 1;
// (+1, +1) and u, v lie in the same tree, then the case number is 2.
// It returns -1 if none of above applies to e.
func (g *Tree) Case(n, m *Node) int {
	return -1
}

// DualUpdate updates the dual values of all blossoms according to their label.
// It adds d if the label is +1, substract d if the label is -1.
func (g *Tree) DualUpdate(d float64) {
	for n, l := range g.label {
		if l == 1 {
			g.dval[n] += d
		} else if l == -1 {
			g.dval[n] -= d
		}
	}
}

// Blossom returns the top blossom id which contains the node n.
func (g *Tree) Blossom(n int64) int64 {
	if _, ok := g.blossom[n]; !ok {
		return n
	}
	return g.Blossom(g.blossom[n])
}

func (g *Tree) Wedges() *Edges {
	var edges []Wedge
	for uid, es := range g.wedges {
		for vid, e := range es {
			if vid < uid {
				continue
			}
			edges = append(edges, e)
		}
	}
	return &Edges{
		pos: -1,
		lst: edges,
	}
}
