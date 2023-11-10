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

// set the blossom (or nodes) as a free node
func (t *Tree) SetFree(b *Node) {
	// fmt.Printf("set free %d (%v)\n", b.temp, b)
	b.label = 0
	b.children = []*Node{}
	b.parent = nil
	for _, c := range b.cycle {
		t.SetFree(c[0].BlossomWithin(b))
	}
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

func (t *Tree) RemoveTight(s [2]*Node) {
	delete(t.tight, s[0])
	delete(t.tight, s[1])
	// for _, u := range s {
	// 	fmt.Printf("delete tight edge on %d\n", u.temp)
	// 	delete(t.tight, u)
	// 	for u.blossom != nil {
	// 		for _, c := range u.blossom.cycle {
	// 			t.RemoveTight(c)
	// 		}
	// 		u = u.blossom
	// 	}
	// }
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
	// fmt.Printf("making cycle...\n")
	// fmt.Printf("anscestary of %d: ", n.temp)
	ansn := n.anscesters()
	// for _, a := range ansn {
	// 	fmt.Printf("%d ", a.temp)
	// }
	// fmt.Println()
	// fmt.Printf("anscestary of %d: ", m.temp)
	ansm := m.anscesters()
	// for _, a := range ansm {
	// 	fmt.Printf("%d ", a.temp)
	// }
	// fmt.Println()
	var i, j int
	for i = 0; i < len(ansn); i++ {
		for j = 0; j < len(ansm); j++ {
			if ansn[i] == ansm[j] {
				goto FOUND
			}
		}
	}
FOUND:
	// fmt.Printf("common parent: %d (i:%d j:%d)\n", ansn[i].temp, i, j)
	var cycle [][2]*Node
	for k := i; k > 0; k-- {
		for l, u := range ansn[k].children {
			ub := u.Blossom()
			if ub == ansn[k-1] {
				ansn[k].children = append(ansn[k].children[:l], ansn[k].children[l+1:]...)
				cycle = append(cycle, [2]*Node{u.parent, u})
				ub.parent = nil
				break
			}
		}
	}
	// fmt.Printf("cycle from parent to n: ")
	// for _, c := range cycle {
	// 	fmt.Printf("%d->%d ", c[0].temp, c[1].temp)
	// }
	// fmt.Println()
	cycle = append(cycle, [2]*Node{n, m})
	// fmt.Printf("append cycle n to m %d->%d\n", n.temp, m.temp)
	for k := 0; k < j; k++ {
		for l, u := range ansm[k+1].children {
			ub := u.Blossom()
			if ub == ansm[k] {
				ansm[k+1].children = append(ansm[k+1].children[:l], ansm[k+1].children[l+1:]...)
				cycle = append(cycle, [2]*Node{u, ub.parent})
				ub.parent = nil
				break
			}
		}
	}
	// fmt.Printf("total cycle: ")
	// for _, c := range cycle {
	// 	fmt.Printf("%d->%d ", c[0].temp, c[1].temp)
	// }
	// fmt.Println()
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
