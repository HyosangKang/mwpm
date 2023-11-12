package mwpm

import (
	"fmt"

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

// set the node n as tight within the blossom b
func (t *Tree) SetTight(s [2]*Node, b *Node) {
	// fmt.Printf("setting tight edge %d:%d\n", s[0].temp, s[1].temp)
	t.tight[s[0]], t.tight[s[1]] = s[1], s[0]
	for _, l := range s {
		for l.blossom != b {
			lb := l.blossom
			for i, c := range lb.cycle {
				if c[0].BlossomWithin(lb) == l.BlossomWithin(lb) {
					lb.cycle = append(lb.cycle[i:], lb.cycle[:i]...)
					break
				}
			}
			for i := 1; i < len(lb.cycle); i += 2 {
				t.SetTight(lb.cycle[i], lb)
			}
			l = l.blossom
		}
	}
}

// remove the tight edge within the blossom b
func (t *Tree) RemoveTight(s [2]*Node) {
	for _, u := range s {
		if _, ok := t.tight[u]; ok {
			delete(t.tight, u)
			t.RemoveTightWithin(u.Blossom())
		}
	}
}

// returns the pair of nodes that connects the cycle.
// For example, it returns [[2 1] [1 0] [0 4] [4 3] [3 2]] if the tree looks like:
//		p o  [2]
//	    /   \
// [1] o     o [3]
//     |     |
// [0] o     o [4]
//	   n --  m

func (t *Tree) makeCycle(n, m *Node) [][2]*Node {
	fmt.Printf("making cycle...\n")
	fmt.Printf("anscestary of %d: ", n.temp)
	ansn := n.anscesters()
	for _, a := range ansn {
		fmt.Printf("%d ", a.temp)
	}
	fmt.Println()
	fmt.Printf("anscestary of %d: ", m.temp)
	ansm := m.anscesters()
	for _, a := range ansm {
		fmt.Printf("%d ", a.temp)
	}
	fmt.Println()
	var i, j int
	for i = 0; i < len(ansn); i++ {
		for j = 0; j < len(ansm); j++ {
			if ansn[i] == ansm[j] {
				goto FOUND
			}
		}
	}
FOUND:
	fmt.Printf("common parent: %d (i:%d j:%d)\n", ansn[i].temp, i, j)
	var cycle [][2]*Node
	for k := i; k > 0; k-- {
		u := ansn[k].popChild(ansn[k-1])
		ub := u.Blossom()
		cycle = append(cycle, [2]*Node{ub.parent, u})
		ub.parent = nil
	}
	fmt.Printf("cycle from parent to n: ")
	for _, c := range cycle {
		fmt.Printf("%d->%d ", c[0].temp, c[1].temp)
	}
	fmt.Println()
	cycle = append(cycle, [2]*Node{n, m})
	fmt.Printf("append cycle n to m %d->%d\n", n.temp, m.temp)
	for k := 0; k < j; k++ {
		u := ansm[k+1].popChild(ansm[k])
		ub := u.Blossom()
		cycle = append(cycle, [2]*Node{u, ub.parent})
		ub.parent = nil
	}
	fmt.Printf("total cycle: ")
	for _, c := range cycle {
		fmt.Printf("%d->%d ", c[0].temp, c[1].temp)
	}
	fmt.Println()
	return cycle
}

func (t *Tree) show() {
	fmt.Printf("tight edges:")
	for u, v := range t.tight {
		fmt.Printf("%d:%d ", u.temp, v.temp)
	}
	fmt.Printf("labels:")
	for _, n := range t.nodes {
		fmt.Printf("%d:%d ", n.temp, n.label)
	}
	fmt.Println()
}

func (t *Tree) TightFrom(n *Node) *Node {
	fmt.Println(n)
	for _, u := range n.all() {
		fmt.Println(t.tight[u])
		if v, ok := t.tight[u]; ok {
			if v.Blossom() != n {
				return v
			}
		}
	}
	panic("No tight match found")
	return nil
}

func (t *Tree) RemoveTightWithin(b *Node) {
	for _, c := range b.cycle {
		for _, u := range c {
			if u.BlossomWithin(b).blossom == b {
				delete(t.tight, u)
			}
		}
		t.RemoveTightWithin(c[0].BlossomWithin(b))
	}
}
