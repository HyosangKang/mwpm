package mwpm

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/graph"
)

// The code takes a weighted undirected graph as input
// and returns a set of perfect matchings that minimizes the sum of weights.
// It is based on Komologov's Blossom V algorithm.
// (We used "multiple trees, constant delta" approach.)
// Technically, a blossom is a set of nodes with cycle, but we call it as a node.
func Run(g graph.Weighted) (map[int64]int64, bool) {
	num := g.Nodes().Len()
	if num%2 == 1 {
		return nil, false
	}
	t := NewTree(g)
	for {
		c, s := t.dualUpdate()
		switch c {
		case 0:
			fmt.Printf("GROW: %d %d\n", s[0].temp, s[1].temp)
			t.grow(s)
		case 1:
			fmt.Printf("AUGMENT: %d %d\n", s[0].temp, s[1].temp)
			t.augment(s)
		case 2:
			fmt.Printf("SHRINK: %d %d\n", s[0].temp, s[1].temp)
			t.shrink(s)
		case 3:
			fmt.Printf("EXPAND: %d\n", s[0].temp)
			t.expand(s[0])
		}
		if len(t.tight) == num {
			break
		}
	}
	inv := make(map[*Node]int64)
	for id, n := range t.nodes {
		inv[n] = id
	}
	new := make(map[int64]int64)
	for n, m := range t.tight {
		new[inv[n]] = inv[m]
	}
	return new, true
}

// dualUpdate updates the dual values of all blossoms according to their label.
// It adds d if the label is +1, substract d if the label is -1.
func (t *Tree) dualUpdate() (int, [2]*Node) {
	var delta, dval float64 = math.MaxFloat64, math.MaxFloat64
	var c int = -1
	var s, y [2]*Node
	for i, n := range t.nodes {
		nb := n.Blossom()
		for j, m := range t.nodes {
			mb := m.Blossom()
			if i >= j || !t.g.HasEdgeBetween(i, j) || nb == mb {
				continue
			}
			slack := t.slack([2]int64{i, j})
			if (nb.label == 0 && mb.label == 1) || (nb.label == 1 && mb.label == 0) { // GROW
				if delta > slack {
					delta = slack
					c = 0
					s = [2]*Node{n, m}
				}
			} else if nb.label == 1 && mb.label == 1 {
				if delta > slack/2 {
					delta = slack / 2
					s = [2]*Node{n, m}
					if nb.Root() != mb.Root() { // AUGMENT
						c = 1
					} else { // SHRINK
						c = 2
					}
				}
			}
			if delta == 0 {
				return c, s
			}
		}
		if nb.label == -1 && (dval > nb.dval) { // EXPAND
			dval = nb.dval
			y = [2]*Node{n, nil}
		}
	}
	if c == -1 { // EXPAND
		delta = dval
		s = y
		c = 3
	}
	/* update dval */
	if delta != 0 {
		for _, n := range t.Blossoms() {
			n.dval += float64(n.label) * delta
		}
	}
	return c, s
}

// returns the slack of an edge e = (u, v) (u, v are not blossoms)
// slack(e) = weight(e) - sum of all dual values of blossoms (including u, v) that containig u and v
func (t *Tree) slack(ids [2]int64) float64 {
	s, _ := t.g.Weight(ids[0], ids[1])
	for _, id := range ids {
		n := t.nodes[id]
		for n != nil {
			s -= n.dval
			n = n.blossom
		}
	}
	return s
}

// GROW makes the tree of u to 'grow'.
//	   o (+) n               o (+) n
//	               --->      |
// (0) o = o (0)         (-) o = o (+)
//	   m   l                 m   l

func (t *Tree) grow(s [2]*Node) {
	n, m := s[0], s[1]
	if n.Blossom().label == 0 {
		n, m = m, n
	}
	l := t.tight[m]
	n, m, l = n.Blossom(), m.Blossom(), l.Blossom()
	/* set parent-child relationship for n and m */
	n.children = append(n.children, m)
	m.parent = n
	/* set parent-child relationship for m and l */
	m.children = []*Node{l}
	l.parent = m
	/* re-label */
	m.label = -1
	l.label = 1
}

// AUGMENT increases the number of matchings.
// (+) o        o (+)                o     o
//	    \     /   \                  I     I
//   (-) o   o (-) o (-)   ----->    o     o   o
//       I   I     I                           I
//   (+) o - o (+) o (+)             o +-+ o   o
//       n   m

func (t *Tree) augment(s [2]*Node) {
	n, m := s[0], s[1]
	/* remove all tight edges from n and m to their roots */
	/* while doing so, set the tight edges */
	for _, l := range s {
		l := l.Blossom()
		for l.parent != nil {
			/* u is the child node and v is the parent node */
			v := l.parent
			var u *Node
			for _, u = range l.parent.Blossom().children {
				if u.Blossom() == l {
					break
				}
			}
			if u.Blossom().label > 0 {
				t.RemoveTight([2]*Node{u, v})
			} else {
				t.SetTight([2]*Node{u, v})
			}
			l = l.parent.Blossom()
		}
		/* set all nodes in two trees as free */
		for _, u := range l.descendents() {
			t.SetFree(u)
		}
	}
	t.SetTight([2]*Node{n, m})
}

// SHRINK makes a new blossom consists of nodes in a tree,
// where two (+) sub-blossoms are connected by edge e.
//	   (-) o p
//	       |                     o p              o (+)
//	   (+) o                     |              /   \
//	     /   \        ---->      o b       (-) o  b  o (-)
//	(-) o     o (-)             / \            |     |
//	    |  e  |            (-) o   o (-)       o  -  o
//	o - o  - o - o                            (+)   (+)
// (-) (+)  (+) (-)                            n     m
//      n    m
// It does not remove the nodes, but there are changes in tree and blossom edges.
// The shrink MUST erase all labels on nodes within.

func (t *Tree) shrink(s [2]*Node) {
	n, m := s[0], s[1]
	/* make a new blossom */
	b := NewNode()
	b.label = 1
	b.cycle = t.makeCycle(n, m) // chain: loop of blossom
	b.parent = b.cycle[0][0].Blossom().parent
	for _, s := range b.cycle {
		sb := s[0].Blossom()
		t.SetFree(sb)
		sb.blossom = b
		b.children = append(b.children, sb.children...)
	}
	for _, c := range b.cycle {
		t.RemoveTight(c)
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
	ansn := n.anscesters()
	ansm := m.anscesters()
	var i, j int
	for i = 0; i < len(ansn); i++ {
		for j = 0; j < len(ansm); j++ {
			if ansn[i] == ansm[j] {
				goto FOUND
			}
		}
	}
FOUND:
	var cycle [][2]*Node
	for k := i; k > 0; k-- {
		for l, u := range ansn[k].children {
			ub := u.Blossom()
			if ub == ansn[k-1] {
				ansn[k].children = append(ansn[k].children[:l], ansn[k].children[l+1:]...)
				ub.parent = nil
				cycle = append(cycle, [2]*Node{u.parent, u})
				break
			}
		}
	}
	cycle = append(cycle, [2]*Node{n, m})
	for k := 0; k < j; k++ {
		for l, u := range ansm[k+1].children {
			ub := u.Blossom()
			if ub == ansm[k] {
				ansm[k+1].children = append(ansm[k+1].children[:l], ansm[k+1].children[l+1:]...)
				ub.parent = nil
				cycle = append(cycle, [2]*Node{u, ub.parent})
				break
			}
		}
	}
	return cycle
}

// EXPAND removes b and add nodes in b to the tree.
// The decision which nodes belong to tree depends on positions of nodes in the cycle.
// For example, if we expand a blossom n consists of [1,2,3], we obtain
//                   [4] o u                     [4] o
//                       |                           |
//  [3] o --- o [2]    n o (-)                   [3] o +-+ o [2]
//      |  n  /          |            ------->            /
//  [1] o --             |                       [1] o --
//                       |                           I
//                   [0] o v                     [0] o
// Nodes that are not added to the tree is matched pairwise.

func (t *Tree) expand(n *Node) {
	/* remove the blossom */
	b := n.Blossom()
	for _, c := range b.cycle {
		sb := c[0].BlossomWithin(b)
		sb.blossom = nil
	}
	/* reorder the cycle start from one that is connected to its parent */
	u := b.parent
	var i int
	for i := 0; i < len(b.cycle); i++ {
		l := b.cycle[i][0].BlossomWithin(b)
		if l.parent == u {
			break
		}
	}
	cycle := append(n.cycle[i:], n.cycle[:i]...)
	chain := append(n.chain[i:], n.chain[:i]...)
	/* find the length of chain in the cycle that goes into the tree */
	i = 0
	if len(b.children) > 0 {
		for j := 2; j < len(cycle); j += 2 {
			sb := cycle[j][0].Blossom()
			if len(sb.children) > 0 {
				i = j
				break
			}
		}
	}
	/* reverse the order of the cycle if necessary */
	if i%2 == 1 {
		rcycle := [][2]*Node{cycle[0]}
		rchain := []*Node{chain[0]}
		for j := len(cycle) - 1; j > 0; j-- {
			rcycle = append(rcycle, cycle[j])
			rchain = append(rchain, chain[j])
		}
		cycle = rcycle
		chain = rchain
	}
	/* make the tree */
	for j := 0; j <= i; j += 2 {
		sb := cycle[j][0].Blossom()
		if j == 0 {
			sb.parent = u
		} else {
			sb.parent = cycle[j-1][0]
		}
		if j == i {
			sb.children = b.children
		} else {
			sb.children = []*Node{cycle[j][1]}
		}
	}
	/* make the tight edges */
	for i := 0; i < len(cycle); i += 2 {
		c := cycle[i]
		u, v := chain[i], chain[i+1]
		t.tight[c[0]], t.tight[c[1]] = c[1], c[0]
		t.MakeTight(c[0], u)
		t.MakeTight(c[1], v)
	}
}
