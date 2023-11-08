package mwpm

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

// MinimumWeightPerfectMatching takes a weighted undirected graph as input
// and returns a set of perfect matchings that minimizes the sum of weights.
// The slice of int64 is the pair of node ids (int64)
// The code is based on Edmond's algorithm using 'primal-dual update,
// explained in Komologov's paper "Blossom V" (2016)
// We use "multiple trees, constant delta" approach.

// The blossom graph does not differentiate nodes and blossoms.
// All nodes are called 'blossom', if stated otherwise.
// (Technically, a blossom means a set of more than one nodes.)
// All blossoms are called by its id(int64).
func Run(g graph.Weighted) (map[int64]int64, bool) {
	num := g.Nodes().Len()
	if num%2 == 1 {
		return nil, false
	}
	t := NewTree(g)
	for {
		// Find a tight edge e. Depending on case, f, do GROW(0), AUGMENT(1), and SHRINK(2).
		// If there is no tight edge, search for a blossom (>3 nodes) with -1 label 0 dual value.
		// If such blossom exists, then do EXPAND(3).
		// If none of above happens, computes the delta, and updates the dual values.
		if nid, mid := t.TightEdge(); nid >= 0 && mid >= 0 {
			n, m := t.nodes[nid].Blossom(), t.nodes[mid].Blossom()
			nl, ml := n.label, m.label
			if (nl == 0 && ml == 1) || (nl == 1 && ml == 0) {
				t.Grow(n, m)
			} else if nl == 1 && ml == 1 {
				if n.Root() != m.Root() {
					t.Augment(n, m)
				} else {
					t.Shrink(n, m)
				}
			}
		} else if n := t.NegBlossom(); n != nil {
			t.Expand(n)
		} else {
			d := t.Delta()
			if math.IsInf(d, +1) || d < 0 {
				break
			}
			t.DualUpdate(t.Delta())
		}
		if len(t.match) == num {
			break
		}
	}
	return t.Match(), true
}

// returns the ids of nodes on a tight edge (slack = 0)
func (t *Tree) TightEdge() (uid, vid int64) {
	for nid, edges := range t.edges {
		for mid, _ := range edges {
			if mid <= nid {
				continue
			}
			if t.slack([2]int64{nid, mid}) == 0 {
				return nid, mid
			}
		}
	}
	return -1, -1
}

// returns the slack of an edge e = (u, v) (u, v are not blossoms)
// slack(e) = weight(e) - sum of all dual values of blossoms (including u, v) that containig u and v
func (g *Tree) slack(ids [2]int64) float64 {
	s := g.edges[ids[0]][ids[1]]
	for i := 0; i < 2; i++ {
		n := g.nodes[ids[i]]
		for n != nil {
			s -= n.dval
			n = n.blossom
		} 
	}
	return s
}

// returns a blossom of -1 label and 0 dual value.
func (t *Tree) NegBlossom() *Node {
	for _, n := range t.Nodes() {
		if n.IsBlossom() && n.label == -1 && n.IsDvalZero() {
			return n
		}
	}
	return nil
}

// 0:GROW, 1:AUGMENT, 2:SHRINK, 3:EXPAND
// Delta returns the minimum value among four types of values:
// slack(u,v) if the edge e=(u,v) is of case 0;
// slack(u,v)/2 if the edge e=(u,v) is of case 1 or 2;
// dualValue(b) if b is a blossom (>3 nodes) of -1 label.
func (t *Tree) Delta() float64 {
	var delta float64 = math.MaxFloat64
	nodes := t.Nodes()
	for i, n := range nodes {
		nb := n.Blossom()
		for j, m := range nodes {
			mb := m.Blossom()
			if (nb.label == 0 & mb.label == 1) || (nb.label == 1 && mb.label == 0) { // GROW
				slack := t.slack([2]int64{i, j})
				if delta > slack {
					delta = slack
				}
			} else if nb.label == 1 && mb.label == 1 {
				slack := t.slack([2]int64{i, j})
				if delta > slack/2 {
					delta = slack / 2
				}
				if nb.Root() != mb.Root() { // AUGMENT
				} else { // SHRINK
				}
			}
		}
		if nb.label == -1 && (delta > nb.dval) { // EXPAND
			delta = nb.dval
		}
	}

		for mid, _ := range edges {
			if mid <= nid {
				continue
			}
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

func (t *Tree) Match() map[int64]int64 {
	inv := make(map[*Node]int64)
	for id, n := range t.nodes {
		inv[n] = id
	}
	new := make(map[int64]int64)
	for n, m := range t.match {
		new[inv[n]] = inv[m]
	}
	return new
}
