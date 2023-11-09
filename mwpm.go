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
	fmt.Println("Making a tree...")
	t := NewTree(g)
	for {
		fmt.Println("Dual update...")
		c, s := t.dualUpdate()
		switch c {
		case 0:
			fmt.Println("GROW")
			t.grow(s)
		case 1:
			fmt.Println("AUGMENT")
			t.augment(s)
		case 2:
			fmt.Println("SHRINK")
			t.shrink(s)
		case 3:
			fmt.Println("EXPAND")
			t.expand(s[0])
		}
		if len(t.tight) == num {
			break
		}
	}
	return t.pair(), true
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
			// fmt.Printf("checking slack between %d and %d: %.2f\n", i, j, slack)
			// fmt.Printf("labels: %d %d\n", nb.label, mb.label)
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
	fmt.Printf("delta: %.2f\n", delta)
	/* update dval */
	if delta != 0 {
		for n := range t.Blossoms() {
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

func (t *Tree) pair() map[int64]int64 {
	inv := make(map[*Node]int64)
	for id, n := range t.nodes {
		inv[n] = id
	}
	new := make(map[int64]int64)
	for n, m := range t.tight {
		new[inv[n]] = inv[m]
	}
	return new
}

// GROW makes the tree of u to 'grow'.
//	   o (+) n               o (+) n
//	               --->      |
// (0) o = o (0)         (-) o = o (+)
//	   m   l                 m   l

func (t *Tree) grow(s [2]*Node) {
	fmt.Printf("%d %d\n", s[0].temp, s[1].temp)
	n, m := s[0], s[1]
	if n.Blossom().label == 0 {
		n, m = m, n
	}
	nb, mb := n.Blossom(), m.Blossom()
	l := t.tight[m]   // node (not blossom) matched to mb
	lb := l.Blossom() // blossom containing l
	/* tree from nb to mb */
	nb.children = append(nb.children, m)
	n.children = append(n.children, m)
	mb.parent = n
	m.parent = n
	/* tree from m to l */
	mb.children = []*Node{l}
	m.children = []*Node{l}
	lb.parent = m
	l.parent = m
	/* relabel */
	mb.label = -1
	lb.label = 1
}

// AUGMENT increases the number of matchings.
// (+) o        o (+)                o     o
//	    \     /   \                  I     I
//   (-) o   o (-) o (-)   ----->    o     o   o
//       I   I     I                           I
//   (+) o - o (+) o (+)             o +-+ o   o
//       n   m

func (t *Tree) augment(s [2]*Node) {
	fmt.Printf("%d %d\n", s[0].temp, s[1].temp)
	n, m := s[0], s[1]
	for _, l := range s {
		lb := l.Blossom()
		for lb.parent != nil {
			var u *Node
			for _, u = range lb.parent.Blossom().children {
				if u.Blossom() == lb {
					break
				}
			}
			v := u.Blossom().parent
			if lb.label > 0 {
				delete(t.tight, u)
				delete(t.tight, v)
			} else {
				t.tight[u], t.tight[v] = v, u
			}
			lb = lb.parent.Blossom()
		}
		r := l.Root()
		for _, u := range r.descendants() {
			t.Free(u)
		}
		delete(t.roots, r)
	}
	t.tight[n], t.tight[m] = m, n
	t.match(n, n.Blossom())
	t.match(m, m.Blossom())
}

// recursively match node n in the blossom b (recursive).
func (t *Tree) match(n, b *Node) {
	if n == b {
		return
	}
	nb := n // nb is the outermost blossom of n within b
	for nb.blossom != b {
		nb = nb.blossom
	}
	/* reorder cycle */
	var i int
	for i = 0; i < len(b.cycle); i++ {
		if b.cycle[i][0] == nb {
			break
		}
	}
	cycle := append(b.cycle[i:], b.cycle[:i]...)
	/* match nodes */
	for i = 1; i < len(cycle); i += 2 {
		t.tight[cycle[i][1]], t.tight[cycle[i][2]] = cycle[i][2], cycle[i][1]
		t.match(cycle[i][1], cycle[i][0])
		t.match(cycle[i][2], cycle[i+1][0])
	}
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

func (t *Tree) shrink(s [2]*Node) {
	fmt.Printf("%d %d\n", s[0].temp, s[1].temp)
	n, m := s[0], s[1]
	nb := NewNode() // new blossom
	nb.label = 1
	nb.cycle = t.cycle(n, m) // cycle, the first is the common parent
	fmt.Printf("CYCLE")
	for _, l := range nb.cycle { // print cycle
		fmt.Printf(" %d->%d ", l[1].temp, l[2].temp)
	}
	nb.parent = nb.cycle[0][0].parent
	for _, ls := range nb.cycle {
		delete(t.tight, ls[1])
		delete(t.tight, ls[2])
		nb.children = append(nb.children, ls[0].children...)
	}
	/* remove nodes in the cycle as children */
	for _, u := range nb.cycle {
		var l *Node
		var i int
		for _, l = range nb.children {
			if l.Blossom() == u[0].Blossom() {
				l = nil
				break
			}
		}
		if l == nil {
			nb.children = append(nb.children[:i], nb.children[i+1:]...)
		}
	}
	for _, ls := range nb.cycle {
		ls[0].label = 0
	}
}

// Cycle finds the common ansester of u, v and cycle of blossoms from p to m and to n.
// For example, it returns [2 4 3 0 1] if the tree looks like:
//		p o  [2]
//	    /   \
// [1] o     o [4]
//     |     |
// [0] o     o [3]
//	   n     m

func (t *Tree) cycle(n, m *Node) [][3]*Node {
	hen := n.anscesters()
	hem := m.anscesters()
	fmt.Println("ANSCESTERS")
	for _, l := range hen {
		fmt.Printf("%d ", l.temp)
	}
	fmt.Println()
	for _, l := range hem {
		fmt.Printf("%d ", l.temp)
	}
	fmt.Println()
	var i1, i2 int
	var nn, mm *Node
	for i1, nn = range hen {
		for i2, mm = range hem {
			if nn.Blossom() == mm.Blossom() {
				goto FOUND
			}
		}
	}
FOUND:
	rev := [][3]*Node{}
	for i := i2; i > 0; i-- {
		rev = append(rev, [3]*Node{hem[i].Blossom(), hem[i], hem[i-1]})
	}
	rev = append(rev, [3]*Node{hem[0].Blossom(), hem[0], n})
	for i := 0; i < i1; i++ {
		rev = append(rev, [3]*Node{hen[i].Blossom(), hen[i], hen[i+1]})
	}
	return rev
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

func (t *Tree) expand(b *Node) {
	/* reorder the cycle start from one that is connected to its parent */
	nb := b.Blossom()
	u := nb.parent
	var i int
	for i := 0; i < len(nb.cycle); i++ {
		if nb.cycle[i][0].parent == nb {
			break
		}
	}
	cycle := append(b.cycle[i:], b.cycle[1:i]...)
	i = len(nb.cycle)
	if len(nb.children) > 0 {
		cb := nb.children[0].Blossom()
		for j, c := range nb.cycle {
			if c[0].children[0].Blossom() == cb {
				i = j
				break
			}
		}
	}
	for j := 0; j < i; j++ {
		if cycle[j][0].label > 0 {
		}
		cycle[j][0].children
	}

	for i = 0; i < len(cycle)-1; i++ {
		l, m := cycle[i], cycle[i+1]
		if i%2 == 0 { // match nodes
			if len(l.Blossom().children) > 0 {
				if l.Blossom().children[0] == v {
					break
				}
			}
			t.tight[l], t.tight[m] = m, l
			l.Blossom().label, m.Blossom().label = -1, 1
		} else {
			l.Blossom().label, m.Blossom().label = 1, -1
		}
		l.Blossom().children = []*Node{m}
		m.Blossom().parent = l
	}
	/* match nodes that are not added to the tree */
	for j := i + 1; j < len(cycle)-1; j += 2 {
		l, m := b.cycle[j], b.cycle[j+1]
		t.tight[l], t.tight[m] = m, l
		t.Free(l)
		t.Free(m)
	}
	/* remove the blossom */
	for _, l := range b.cycle {
		for l.blossom != nb {
			l = l.blossom
		}
		l.blossom = nil
	}
}
