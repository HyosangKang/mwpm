package mwpm

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"gonum.org/v1/gonum/graph"
)

// The code takes a weighted undirected graph as input
// and returns a set of perfect matchings that minimizes the sum of weights.
// It is based on Komologov's Blossom V algorithm.
// (We used "multiple trees, constant delta" approach.)
// Technically, a blossom is a set of nodes with cycle, but we call it as a node.
func Run(g graph.Weighted) ([][2]int64, bool) {
	num := g.Nodes().Len()
	if num%2 == 1 {
		return nil, false
	}
	t := NewTree(g)
	for {
		if len(t.tight) == num {
			break
		}
		c, s := t.dual()
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
			fmt.Println("EXPAND")
			fmt.Printf("EXPAND: %d\n", s[0].temp)
			t.expand(s[0])
		}
	}
	for _, m := range t.tight {
		if _, ok := t.tight[m]; !ok {
			panic("NO TIGHT PAIR")
		}
	}
	inv := make(map[*Node]int64)
	for id, n := range t.nodes {
		inv[n] = id
	}
	var match [][2]int64
	for n, m := range t.tight {
		i, j := inv[n], inv[m]
		if i < j {
			match = append(match, [2]int64{inv[n], inv[m]})
		}
	}
	sort.Slice(match, func(i, j int) bool {
		return match[i][0] < match[j][0]
	})
	return match, true
}

// dual updates the dual values of all blossoms according to their label.
// It adds d if the label is +1, substract d if the label is -1.
func (t *Tree) dual() (int, [2]*Node) {
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
			if (nb.label == 0 && mb.label > 0) && delta > slack {
				delta = slack
				c = 0
				s = [2]*Node{m, n}
			} else if (nb.label > 0 && mb.label == 0) && delta > slack { // GROW
				delta = slack
				c = 0
				s = [2]*Node{n, m}
			} else if (nb.label > 0 && mb.label > 0) && delta > slack/2 {
				delta = slack / 2
				s = [2]*Node{n, m}
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
		if nb.label < 0 && (dval > nb.dval) { // EXPAND
			dval = nb.dval
			y = [2]*Node{n, nil}
		}
	}
	if c == -1 { // EXPAND
		delta = dval
		s = y
		c = 3
	}
	fmt.Printf("delta: %f\n", delta)
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
//	n (+)                   (+)
//	               --->      |
//  m (0) = (0)             (-) = (+)
//	   u     v
// tight pair: (u, v), u is in the blossom of m

func (t *Tree) grow(s [2]*Node) {
	n, m := s[0], s[1]
	nb, mb := n.Blossom(), m.Blossom()
	v := t.TightFrom(mb)
	vb := v.Blossom()
	u := t.TightFrom(vb)
	if m.Blossom() != u.Blossom() {
		panic("Error in GROW: m.blossom != u.blossom")
	}
	/* set parent-child relationship for n and m */
	nb.children = append(nb.children, m)
	mb.parent = n
	/* set parent-child relationship for m and l */
	mb.children = []*Node{v}
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

func (t *Tree) augment(s [2]*Node) {
	/* remove all tight edges from n and m to their roots */
	/* while doing so, set the tight edges */
	for _, l := range s {
		pairs := l.anscestary()
		for i := 0; i < len(pairs); i += 2 {
			// fmt.Printf("Calling RemoveTight from AUGMENT %d, %d\n", p[0].temp, p[1].temp)
			t.RemoveTight(pairs[i])
		}
		for i := 1; i < len(pairs); i += 2 {
			t.SetTight(pairs[i], nil)
		}
		for _, u := range l.Root().descendents() {
			u.SetFree()
		}
	}
	t.SetTight(s, nil)
}

// SHRINK makes a new blossom consists of nodes in a tree,
// where two (+) sub-blossoms are connected by edge e.
//	        (-) p
//	         I                    (-) p            (+)
//	        (+)                    I              /   \
//	       /   \        ---->     (+) b         (-) b (-)
//	     (-)   (-)                / \            I     I
//	      I  e  I               (-) (-)         (+) - (+)
// (-) - (+) - (+) - (-)                         n     m
//        n     m

func (t *Tree) shrink(s [2]*Node) {
	n, m := s[0], s[1]
	/* make a new blossom */
	b := &Node{
		temp:  rand.Int63(),
		label: 1,
		cycle: t.makeCycle(n, m), // chain: loop of blossom
	}
	b.parent = b.cycle[0][0].Blossom().parent
	for _, c := range b.cycle {
		cb := c[0].Blossom()
		b.dval += float64(cb.label) * cb.dval
		b.children = append(b.children, cb.children...)
		cb.blossom = b
	}
}

// EXPAND removes b and add nodes in b to the tree.
// The decision which nodes belong to tree depends on positions of nodes in the cycle.
// For example, if we expand a blossom n consists of [1,2,3], we obtain
//                  [4] (+) u                   [4] (+)
//                       |                           |
//  [3] o --- o [2]   n (-)                     [3] (-) +-+ (+) [2]
//      |  n  /          I            ------->             /
//  [1] o --        [0] (+) v                   [1] (-) --
//                                                   I
//                                              [0] (+)
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
		rcycle := [][2]*Node{}
		for j := len(b.cycle) - 1; j > 0; j-- {
			rcycle = append(rcycle, [2]*Node{rcycle[j][1], rcycle[j][0]})
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
			sb.children = []*Node{b.cycle[j][1]}
		}
		if j%2 == 0 {
			sb.label = -1
		} else {
			sb.label = 1
		}
	}
	/* set free the rest */
	for j := i + 1; j < len(b.cycle); j++ {
		b.cycle[j][0].Blossom().SetFree()
	}
	/* make the tight edges */
	for j := 0; j < len(b.cycle); j += 2 {
		t.SetTight(b.cycle[j], nil)
	}
}
