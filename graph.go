package mwpm

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/graph"
)

var _ graph.Weighted = (*Graph)(nil)

// Graph consists of
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
type Graph struct {
	nodes  map[int64]Node
	bedges map[int64]map[int64]Bedge
	wedges map[int64]map[int64]Wedge // contains info on original edges
}

func (g *Graph) Node(id int64) graph.Node {
	return g.nodes[id]
}

func (g *Graph) Nodes() graph.Nodes {
	var lst []Node
	for _, n := range g.nodes {
		lst = append(lst, n)
	}
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].ID() < lst[j].ID()
	})
	return &Nodes{pos: -1, lst: lst}
}

func (g *Graph) From(id int64) graph.Nodes {
	var lst []Node
	for id := range g.edges[id] {
		lst = append(lst, g.nodes[id])
	}
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].ID() < lst[j].ID()
	})
	return &Nodes{pos: -1, lst: lst}
}

func (g *Graph) HasEdgeBetween(xid, yid int64) bool {
	_, ok := g.Weight(xid, yid)
	return ok
}

func (g *Graph) Edge(uid, vid int64) graph.Edge {
	e, ok := g.edges[uid][vid]
	if !ok {
		return nil
	}
	return e
}

func (g *Graph) WeightedEdge(uid, vid int64) graph.WeightedEdge {
	e, ok := g.edges[uid][vid]
	if !ok {
		return nil
	}
	return e
}

func (g *Graph) Weight(xid, yid int64) (float64, bool) {
	e, ok := g.edges[xid][yid]
	w := e.Weight()
	return w, ok
}

// NewBlossomGraphFrom takes a weighted undirected graph as input
// and retuns a blossom graph with initialization.
// The blossom graph contains only nodes, and no blossom edge.
// Each node becomes a blossom of itself.
// The labels are set to +1.
// The root, parent of each blossom is itself and empty children.
func GraphFrom(wg graph.Weighted) *Graph {
	g := &Graph{
		nodes: make(map[int64]Node),
		edges: make(map[int64]map[int64]Bedge),
	}
	nodes := wg.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		n := NewNode()
		n.id = id
		g.AddNode(n)
	}
	nodes = g.Nodes()
	for nodes.Next() {
		n := nodes.Node().(Node)
		nid := n.ID()
		modes := g.Nodes()
		for modes.Next() {
			m := modes.Node().(Node)
			if n.ID() >= m.ID() {
				continue
			}
			mid := m.ID()
			g.AddEdge(Bedge{
				nodes:  [2]Node{n, m},
				weight: wg.WeightedEdge(nid, mid).Weight(),
			})
		}
	}
	return g
}

func (g *Graph) AddEdge(e Bedge) {
	u := e.From().ID()
	v := e.To().ID()
	if _, ok := g.edges[u]; !ok {
		g.edges[u] = make(map[int64]Bedge)
	}
	g.edges[u][v] = e
	if _, ok := g.edges[v]; !ok {
		g.edges[v] = make(map[int64]Bedge)
	}
	g.edges[v][u] = e.ReversedEdge().(Bedge)
}

func (g *Graph) newID() int64 {
	var id int64 = 0
	for {
		if _, ok := g.nodes[id]; !ok {
			break
		}
		id++
	}
	return id
}

func (g *Graph) AddNode(n Node) {
	id := g.newID()
	n.cycle = []int64{id}
	n.nodes = []int64{id}
	g.nodes[id] = n
}

// Match returns all pairs of matched nodes.
// It counts the same match twice, only for simplicity
func (g *Graph) Match() [][2]int64 {
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
func (g *Graph) TightEdge() (Wedge, int8) {
	var we Wedge
	var rc int8
	edges := g.Wedges()
	for edges.Next() {
		e := edges.Edge().(Wedge)
		s := g.Slack(e)
		if s == 0 {
			c := g.Case(e)
			if c == 1 {
				return e, 1
			}
			if c == 0 {
				we = e
				rc = c
			}
		}
	}
	return we, rc
}

// NegBlossom returns a blossom (>3 nodes) with -1 label and 0 dual value.
// It returns -1 if no such blossom exists.
func (g *Graph) NegBlossom() int64 {
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
func (g *Graph) Delta() float64 {
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
func (g *Graph) Slack(e graph.WeightedEdge) float64 {
	s := e.Weight()
	nodes := []Node{g.Node(e.From().ID()).(Node), g.Node(e.To().ID()).(Node)}
	for _, n := range nodes {
		if n.IsBlossom() {
			for _, mid := range n.nodes {
				m := g.Node(mid).(Node)
				s -= m.dval
			}
		} else {
			s -= n.DualVal()
		}
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
func (g *Graph) Case(e graph.WeightedEdge) int8 {
	if e.From().ID() == e.To().ID() {
		return -1
	}
	n := g.Node(e.From().ID()).(Node)
	m := g.Node(e.To().ID()).(Node)
	nl, ml := n.Label(), m.Label()
	if nl*ml == 0 && nl+ml == 1 {
		return 0
	} else if nl*ml == 1 {
		if n.Root() != m.Root() {
			return 1
		} else {
			return 2
		}
	}
	return -1
}

// DualUpdate updates the dual values of all blossoms according to their label.
// It adds d if the label is +1, substract d if the label is -1.
func (g *Graph) DualUpdate(d float64) {
	for n, l := range g.label {
		if l == 1 {
			g.dval[n] += d
		} else if l == -1 {
			g.dval[n] -= d
		}
	}
}

// Blossom returns the top blossom id which contains the node n.
func (g *Graph) Blossom(n int64) int64 {
	if _, ok := g.blossom[n]; !ok {
		return n
	}
	return g.Blossom(g.blossom[n])
}

func (g *Graph) Wedges() *Edges {
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
