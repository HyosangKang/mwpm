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
		// t.show()
		if len(t.tight) == num {
			break
		}
	}
	inv := make(map[*Blossom]int64)
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
func (t *Tree) dualUpdate() (int, [2]*Blossom) {
	var delta, dval float64 = math.MaxFloat64, math.MaxFloat64
	var c int = -1
	var s, y [2]*Blossom
	for i, n := range t.nodes {
		nb := n.Blossom()
		for j, m := range t.nodes {
			mb := m.Blossom()
			if i >= j || !t.g.HasEdgeBetween(i, j) || nb == mb {
				continue
			}
			slack := t.slack([2]int64{i, j})
			if (nb.label == 0 && mb.label == 1) && delta > slack {
				delta = slack
				c = 0
				s = [2]*Blossom{m, n}
			} else if (nb.label == 1 && mb.label == 0) && delta > slack { // GROW
				delta = slack
				c = 0
				s = [2]*Blossom{n, m}
			} else if (nb.label == 1 && mb.label == 1) && delta > slack/2 {
				delta = slack / 2
				s = [2]*Blossom{n, m}
				if nb.Root() != mb.Root() { // AUGMENT
					c = 1
				} else { // SHRINK
					c = 2
				}
			}
			if delta == 0 {
				return c, s
			}
		}
		if nb.label == -1 && (dval > nb.dval) { // EXPAND
			dval = nb.dval
			y = [2]*Blossom{n, nil}
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
			n.update(delta)
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
//	   m                     m
// tight pair: (u, v), u is in the blossom of m

func (t *Tree) grow(s [2]*Blossom) {
	n, m := s[0], s[1]
	nb, mb := n.Blossom(), m.Blossom()
	v := t.TightWith(mb)
	vb := v.Blossom()
	u := t.TightWith(vb)
	/* set parent-child relationship for n and m */
	nb.children = append(nb.children, m)
	mb.parent = n
	/* set parent-child relationship for m and l */
	mb.children = []*Blossom{v}
	vb.parent = u //
	/* re-label */
	mb.label = -1
	vb.label = 1
}

// AUGMENT increases the number of matchings.
// (+) o        o (+)                o     o
//	    \     /   \                  I     I
//   (-) o   o (-) o (-)   ----->    o     o   o
//       I   I     I                           I
//   (+) o - o (+) o (+)             o +-+ o   o
//       n   m

func (t *Tree) augment(s [2]*Blossom) {
	n, m := s[0], s[1]
	/* remove all tight edges from n and m to their roots */
	/* while doing so, set the tight edges */
	for _, l := range s {
		// fmt.Printf("resetting tight edges from %d\n", l.temp)
		cycle := l.anscestary()
		for _, c := range cycle {
			if c[0].Blossom().label > 0 {
				t.RemoveTight(c, nil)
			}
		}
		for _, c := range cycle {
			if c[0].Blossom().label < 0 {
				t.SetTight(c, nil)
			}
		}
		for _, u := range l.Root().descendents() {
			t.SetFree(u)
		}
	}
	t.SetTight([2]*Blossom{n, m}, nil)
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

func (t *Tree) shrink(s [2]*Blossom) {
	n, m := s[0], s[1]
	/* make a new blossom */
	b := NewNode()
	b.label = 1
	b.cycle = t.makeCycle(n, m) // chain: loop of blossom
	b.parent = b.cycle[0][0].Blossom().parent
	for _, c := range b.cycle {
		cb := c[0].Blossom()
		b.dval += cb.dval
		b.children = append(b.children, cb.children...)
		cb.blossom = b
		cb.SetAlone()
	}
	for _, c := range b.cycle {
		if c[0].Blossom() != b {
			panic("blossom is not set correctly")
		}
	}
	for _, c := range b.cycle {
		t.RemoveTight(c, b)
	}
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

func (t *Tree) expand(n *Blossom) {
	/* remove the blossom */
	b := n.Blossom()
	for _, c := range b.cycle {
		sb := c[0].BlossomWithin(b)
		sb.blossom = nil
	}
	/* reorder the cycle start from one that is connected to its parent */
	u := b.parent
	for i, c := range b.cycle {
		for _, l := range u.children {
			if l.Blossom() == c[0].Blossom() {
				b.cycle = append(b.cycle[i:], b.cycle[:i]...)
				goto FOUND
			}
		}
	}
FOUND:
	/* find the length of chain in the cycle that goes into the tree */
	i := len(b.cycle) - 1
	if len(b.children) > 0 {
		v := b.children[0].Blossom().parent.Blossom()
		for j, c := range b.cycle {
			if c[0].Blossom() == v {
				i = j
				break
			}
		}
	}
	/* reverse the order of the cycle if necessary */
	if i%2 == 1 {
		rcycle := [][2]*Blossom{}
		for j := len(b.cycle) - 1; j > 0; j-- {
			rcycle = append(rcycle, [2]*Blossom{rcycle[j][1], rcycle[j][0]})
		}
		b.cycle = rcycle
		i = len(b.cycle) - i
	}
	/* make the tree */
	for j := 0; j <= i; j += 2 {
		sb := b.cycle[j][0].Blossom()
		if j == 0 {
			sb.parent = u
		} else {
			sb.parent = b.cycle[j-1][0]
		}
		if j == i {
			sb.children = b.children
		} else {
			sb.children = []*Blossom{b.cycle[j][1]}
		}
		if j%2 == 0 {
			sb.label = -1
		} else {
			sb.label = 1
		}
	}
	/* make the tight edges */
	for i := 0; i < len(b.cycle); i += 2 {
		t.SetTight(b.cycle[i], nil)
	}
}
