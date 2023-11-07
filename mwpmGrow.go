package mwpm

// Grow takes an weighted edge e to attach a blossom edge to the blossom graph.
// The opration is depicted as
//	o (+) n                o (+) n
//	               --->     \
// (0) o = o (0)         (-) o = o (+)
//	    m                     m
// Grow makes the tree of u to 'grow'.
// No new match is made in Grow.
// The fields that needs updates are:
// 0. labels of v, w
// 1. edges bewteen u, v
// 2. parent, root, children of u, v, w

func (t *Tree) Grow(n, m *Node) {
	if n.label == 0 {
		n, m = m, n
	}
	n.children[m] = struct{}{}
	m.label = -1
	m.parent = n
	m.children[n] = struct{}{}
	t.match[m].label = 1
	t.match[m].parent = m
}
