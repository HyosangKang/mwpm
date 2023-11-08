package mwpm

import (
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
		c, s := t.DualUpdate()
		switch c {
		case 0:
			t.grow(s)
		case 1:
			t.augment(s)
		case 2:
			t.shrink(s)
		case 3:
			t.expand(s[0])
		}
		if len(t.tight) == num {
			break
		}
	}
	return t.match(), true
}

// DualUpdate updates the dual values of all blossoms according to their label.
// It adds d if the label is +1, substract d if the label is -1.
func (t *Tree) DualUpdate() (int, [2]*Node) {
	var delta, dval float64 = math.MaxFloat64, math.MaxFloat64
	var c int = -1
	var s, y [2]*Node
	for i, n := range t.nodes {
		nb := n.Blossom()
		for j, m := range t.nodes {
			mb := m.Blossom()
			if (nb.label == 0 && mb.label == 1) || (nb.label == 1 && mb.label == 0) { // GROW
				slack := t.slack([2]int64{i, j})
				if delta > slack {
					delta = slack
					c = 0
					s = [2]*Node{n, m}
				}
			} else if nb.label == 1 && mb.label == 1 {
				slack := t.slack([2]int64{i, j})
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
	if math.Abs(delta) > Eps {
		for n := range t.Blossoms() {
			n.dval += float64(n.label) * delta
		}
	}
	return c, s
}

// returns the slack of an edge e = (u, v) (u, v are not blossoms)
// slack(e) = weight(e) - sum of all dual values of blossoms (including u, v) that containig u and v
func (g *Tree) slack(ids [2]int64) float64 {
	s := g.edges[ids[0]][ids[1]]
	for _, id := range ids {
		n := g.nodes[id]
		for n != nil {
			s -= n.dval
			n = n.blossom
		}
	}
	return s
}

func (t *Tree) match() map[int64]int64 {
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
	n, m := s[0], s[1]
	nb, mb := n.Blossom(), m.Blossom()
	if nb.label == 0 {
		n, m = m, n
	}
	l := t.tight[m]   // node (not blossom) matched to mb
	lb := l.Blossom() // blossom containing l
	/* tree from nb to mb */
	nb.children = append(nb.children, m)
	mb.parent = n
	mb.children = []*Node{l}
	n.children = []*Node{m}
	m.parent = n
	/* tree from m to l */
	mb.children = []*Node{l}
	lb.parent = m
	m.children = []*Node{l}
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
	n, m := s[0], s[1]
	t.tight[n], t.tight[m] = m, n
	for _, l := range s {
		for u := l; u.parent != nil; u = u.parent {
			if u.Blossom().label < 0 {
				t.tight[u], t.tight[u.parent] = u.parent, u
			}
		}
		r := l.Root()
		for _, v := range r.descendants() {
			t.Free(v.Blossom())
		}
		delete(t.roots, r)
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
	n, m := s[0], s[1]
	nn := NewNode()
	nn.label = 1
	nn.cycle = t.cycle(n, m)       // cycle, the first is the common parent
	nn.parent = nn.cycle[0].parent // grandparent is now new parent
	for _, l := range nn.cycle[1:] {
		if len(l.children) > 0 {
			nn.cycle[0].children = append(nn.cycle[0].children, l.children...)
		}
	}
	for _, l := range nn.cycle {
		l.label = 0
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

func (t *Tree) cycle(n, m *Node) []*Node {
	hen := n.anscesters()
	hem := m.anscesters()
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
	rev := []*Node{}
	for i := i2; i >= 0; i-- {
		rev = append(rev, hem[i])
	}
	return append(rev, hen[:i1]...)

}

// EXPAND removes b and add nodes in b to the tree.
// The decision which nodes belong to tree depends on positions of nodes in the cycle.
// For example, if we expand a blossom n consists of [1,2,3], we obtain
//                   [4] o                       [4] o
//                       |                           |
//  [3] o --- o [2]    n o                       [3] o +-+ o [2]
//      |  n  /          |            ------->            /
//  [1] o --             |                       [1] o --
//                       |                           I
//                   [0] o                       [0] o
// Nodes that are not added to the tree is matched pairwise.

func (t *Tree) expand(n *Node) {
	/* expand tree */
	var i int
	for i = 0; i < len(n.cycle)-1; i++ {
		l, m := n.cycle[i], n.cycle[i+1]
		l.Blossom().children = []*Node{m}
		m.Blossom().parent = l
		if i%2 == 0 { // match nodes
			t.tight[l], t.tight[m] = m, l
			l.Blossom().label = -1
		} else {
			l.Blossom().label = 1
		}
		if len(l.Blossom().children) > 0 {
			break
		}
	}
	/* match nodes that are not added to the tree */
	for j := i + 1; j < len(n.cycle)-1; j += 2 {
		l, m := n.cycle[j], n.cycle[j+1]
		t.tight[l], t.tight[m] = m, l
		t.Free(l.Blossom())
		t.Free(m.Blossom())
	}
	/* remove the blossom */
	for _, l := range n.cycle {
		for l.blossom != n {
			l = l.blossom
		}
		l.blossom = nil
	}
}
