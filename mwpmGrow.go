package mwpm

// Grow makes the tree of u to 'grow'
// The opration is depicted as
//	   o (+) n               o (+) n
//	               --->      |
// (0) o = o (0)         (-) o = o (+)
//	   m   l                 m   l

func (t *Tree) Grow(n, m *Node) {
	if n.label == 0 {
		n, m = m, n
	}
	n.children = append(n.children, m)
	m.label = -1
	m.parent = n
	l := t.match[m]
	m.children = []*Node{l}
	l.label = 1
	l.parent = m
}
