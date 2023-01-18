package mwpm

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
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
func MinimumWeightPerfectMatching(wg *simple.WeightedUndirectedGraph, msg bool) ([][2]int64, bool) {

	// Initialize the blossom graph. See the description of the function.
	g := NewBlossomGraphFrom(wg)

	// Find the total number of nodes (must be even).
	num := 0
	all := wg.Nodes()
	for all.Next() {
		num++
	}

	// If there are odd number of nodes, abort.
	match := make([][2]int64, 0, num)
	if num == 0 {
		return match, true
	}
	if num%2 == 1 {
		fmt.Printf("There are odd nubmer of nodes. Abort the matching.\n")
		return match, false
	}

	// expanded := false
	// when := 0
	loop := 0
	for true {

		// Find a tight edge e. Depending on case, f, do GROW(0), AUGMENT(1), and SHRINK(2).
		// If there is no tight edge, search for a blossom (>3 nodes) with -1 label 0 dual value.
		// If such blossom exists, then do EXPAND(3).
		// If none of above happens, computes the delta, and updates the dual values.
		if e, c := g.TightEdge(); c > -1 {
			if msg {
				fmt.Printf("[%d] A tight edge %+v is found.\n", loop, e)
			}
			switch c {
			case 0:
				if msg {
					fmt.Printf("=============GROW===============\n")
				}
				g.Grow(e)
			case 1:
				if msg {
					fmt.Printf("============AUGMENT=============\n")
				}
				g.Augment(e)
			case 2:
				if msg {
					fmt.Printf("============SHRINK==============\n")
				}
				g.Shrink(e)
			}
		} else if b := g.NegBlossom(); b > -1 {
			if msg {
				fmt.Printf("A negative blossom [%+v:%+v] found.\n", b, g.cycle[b])
				fmt.Printf("============EXPAND==============\n")
			}
			// expanded = true
			// when = loop
			g.Expand(b)
		} else {
			d := g.Delta()
			if msg {
				fmt.Printf("No tight edges nor negative blossom found. %.1f delta update.\n", d)
			}
			if math.IsInf(d, +1) || d < 0 {
				if msg {
					fmt.Printf("No feasible edge/blossom found. Abort the matching.\n")
				}
				break
			}
			g.DualUpdate(g.Delta())
		}

		// Abort loop if number of matched nodes equal to the total nubmer of nodes.
		// This happends only when all nodes are matched
		// because matching only counts between nodes, not blossoms.
		match = g.Match()
		if msg {
			fmt.Printf("Label: %+v\n", g.label)
			fmt.Printf("Blossom: %+v\n", g.blossom)
			fmt.Printf("Nodes: %+v\n", g.nodes)
			fmt.Printf("Root: %+v\n", g.root)
			fmt.Printf("Parent: %+v\n", g.parent)
			fmt.Printf("Children: %+v\n", g.children)
			fmt.Printf("Dual value: %+v\n", g.dval)
			fmt.Printf("Cycle: %+v\n", g.cycle)
			fmt.Printf("Match: %+v\n", match)
			fmt.Printf("================================\n\n")
		}
		if len(match) == num/2 {
			break
		}

		loop++
		// For debugging purpose only.
		// if loop > 100 {
		// 	fmt.Printf("Abort due to too many operations.\n")
		// 	break
		// }
	}

	// if expanded {
	// fmt.Printf("EXPAND occured at %d.\n", when)
	// }
	return match, true
}

// A BlossomEdge consists of WeightedEdge and isMatch(bool)
// Two blossoms should be connected by a BlossomEdge only.
type BlossomEdge struct {
	e     graph.WeightedEdge
	match bool
}

// BlossomGraph consists of
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
type BlossomGraph struct {
	wg                     *simple.WeightedUndirectedGraph
	label                  map[int64]int
	root, parent, blossom  map[int64]int64
	children, cycle, nodes map[int64][]int64
	dval                   map[int64]float64
	edges                  map[int64]map[int64]BlossomEdge
}

// NewBlossomGraphFrom takes a weighted undirected graph as input
// and retuns a blossom graph with initialization.
// The blossom graph contains only nodes, and no blossom edge.
// Each node becomes a blossom of itself.
// The labels are set to +1.
// The root, parent of each blossom is itself and empty children.
func NewBlossomGraphFrom(wg *simple.WeightedUndirectedGraph) *BlossomGraph {
	g := BlossomGraph{
		wg:       wg,
		label:    make(map[int64]int),
		root:     make(map[int64]int64),
		parent:   make(map[int64]int64),
		blossom:  make(map[int64]int64),
		children: make(map[int64][]int64),
		cycle:    make(map[int64][]int64),
		nodes:    make(map[int64][]int64),
		dval:     make(map[int64]float64),
		edges:    make(map[int64]map[int64]BlossomEdge),
	}
	all := wg.Nodes()
	for all.Next() {
		id := int64(len(g.label))
		g.label[id] = +1
		g.root[id] = id
		g.parent[id] = id
		g.cycle[id] = []int64{id}
		g.nodes[id] = []int64{id}
		g.dval[id] = 0
	}
	return &g
}

// Match returns all pairs of matched nodes.
// It counts the same match twice, only for simplicity
func (g *BlossomGraph) Match() [][2]int64 {
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
// An edge e is called tight if slack(e) = 0 (See Slack for formula.)
// Since augment occurs when the case number is 1, it has priority among others.
func (g *BlossomGraph) TightEdge() (graph.WeightedEdge, int) {
	var re graph.WeightedEdge
	rc := -1

	all := g.wg.WeightedEdges()
	for all.Next() {
		e := all.WeightedEdge()
		s := g.Slack(e)
		if s == 0 {
			c := g.Case(e)
			if c == 1 {
				return e, 1
			} else if c > -1 {
				re = e
				rc = c
			}
		}
	}
	return re, rc
}

// NegBlossom returns a blossom (>3 nodes) with -1 label and 0 dual value.
// It returns -1 if no such blossom exists.
func (g *BlossomGraph) NegBlossom() int64 {
	for n, l := range g.label {
		if l == -1 {
			if len(g.cycle[n]) > 1 && g.dval[n] == 0 {
				return n
			}
		}
	}
	return -1
}

// Delta returns the minimum value among four types of values:
// slack(u,v) if the edge e=(u,v) is of case 0;
// slack(u,v)/2 if the edge e=(u,v) is of case 1 or 2;
// dualValue(b) if b is a blossom (>3 nodes) of -1 label.
func (g *BlossomGraph) Delta() float64 {
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
func (g *BlossomGraph) Slack(e graph.WeightedEdge) float64 {
	s := e.Weight()
	u := e.From().ID()
	v := e.To().ID()

	for i, f := range g.nodes {
		if ContainsInt64(f, u) || ContainsInt64(f, v) {
			s -= g.dval[i]
		}
	}
	// return e.Weight() - g.dval[e.From().ID()] - g.dval[e.To().ID()]
	return s
}

// Case returns the case of an edge e = (u, v) as below
// Let n, m be the top blossom containing u, v respectively.
// If the labels of n, m are
// (+1, 0) or (0, +1), then the case number is 0;
// (+1, +1) and u, v lie in different tree, then the case number is 1;
// (+1, +1) and u, v lie in the same tree, then the case number is 2.
// It returns -1 if none of above applies to e.
func (g *BlossomGraph) Case(e graph.WeightedEdge) int {
	n := g.Blossom(e.From().ID())
	m := g.Blossom(e.To().ID())

	if n == m {
		return -1
	}
	label := [2]int{g.label[n], g.label[m]}
	if label == [2]int{1, 0} || label == [2]int{0, 1} {
		return 0
	} else if label == [2]int{1, 1} {
		if g.root[n] != g.root[m] {
			return 1
		} else {
			return 2
		}
	}
	return -1
}

// DualUpdate updates the dual values of all blossoms according to their label.
// It adds d if the label is +1, substract d if the label is -1.
func (g *BlossomGraph) DualUpdate(d float64) {
	for n, l := range g.label {
		if l == 1 {
			g.dval[n] += d
		} else if l == -1 {
			g.dval[n] -= d
		}
	}
}

// Blossom returns the top blossom id which contains the node n.
func (g *BlossomGraph) Blossom(n int64) int64 {
	if _, ok := g.blossom[n]; !ok {
		return n
	}
	return g.Blossom(g.blossom[n])
}
