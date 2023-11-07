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
func MinimumWeightPerfectMatching(wg graph.Weighted) ([][2]int64, bool) {
	if wg.Nodes().Len()%2 == 1 {
		return nil, false
	}
	g := GraphFrom(wg)
	loop := 0
	for true {
		// Find a tight edge e. Depending on case, f, do GROW(0), AUGMENT(1), and SHRINK(2).
		// If there is no tight edge, search for a blossom (>3 nodes) with -1 label 0 dual value.
		// If such blossom exists, then do EXPAND(3).
		// If none of above happens, computes the delta, and updates the dual values.
		if e, c := g.TightEdge(); c > -1 {
			switch c {
			case 0:
				g.Grow(e)
			case 1:
				g.Augment(e)
			case 2:
				g.Shrink(e)
			}
		} else if b := g.NegBlossom(); b > -1 {
			g.Expand(b)
		} else {
			d := g.Delta()
			if math.IsInf(d, +1) || d < 0 {
				break
			}
			g.DualUpdate(g.Delta())
		}
		// Abort loop if number of matched nodes equal to the total nubmer of nodes.
		// This happends only when all nodes are matched
		// because matching only counts between nodes, not blossoms.
		match = g.Match()
		if len(match) == num/2 {
			break
		}

		loop++
	}
	return match, true
}

// A BlossomEdge consists of WeightedEdge and isMatch(bool)
// Two blossoms should be connected by a BlossomEdge only.
