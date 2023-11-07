package mwpm

// Shrink makes a new blossom consists of nodes in a tree,
// where two (+) sub-blossoms are connected by edge e.
//	   (-) o p
//	       |                     o p              o (+)
//	   (+) o                     |              /   \
//	     /   \        ---->      o b       (-) o  b  o (-)
//	(-) o     o (-)             / \            |     |
//	    |  e  |            (-) o   o (-)       o  -  o
//	o - o  - o - o                            (+)   (+)
// (-) (+)  (+) (-)
// It does not remove the nodes, but there are changes in tree and blossom edges.

func (t *Tree) Shrink(n, m *Node) {
	// Find the common anscester of u and v.
	// Then create a cycle of blossoms in the followwing order: u -> ... -> p -> ... -> v
	// Also create edge between u and v
	nn := NewNode()
	cycle := t.Cycle(n, m)             // cycle, the first is the common parent
	comm := cycle[0]                   // common parent of n, m
	nn.parent = comm.parent            // grandparent is now new parent
	delete(comm.parent.children, comm) // remove the child (the parent of n, m) from grandparent
	comm.children[nn] = struct{}{}
	nn.cycle = cycle
	for _, c := range cycle {
		c.label = 0
	}

	// Create new blossom b consists of blossoms in the cycle.
	// The label of b is set to +1, and dual value 0.
	// The field blossom, nodes are initialized set as well.
	b := int64(len(t.label))
	t.label[b] = +1
	t.cycle[b] = cycle
	t.nodes[b] = []int64{}
	for _, n := range cycle {
		t.nodes[b] = append(t.nodes[b], t.nodes[n]...)
	}
	t.dval[b] = 0

	// The blossoms in the cycle now belong to the blossom b.
	// All labels are set to 0.
	for _, n := range cycle {
		t.blossom[n] = b
		t.LabelAsZero(n)
	}

	// If p is not the root of the tree, remove the edges between parent of p (pp) and p,
	// and create a edge between b and pp.
	// Remove all edges between blossoms in the cycle and their children,
	// and replace by edges between b and the children
	pp := t.parent[p]
	if pp != p {
		be := t.edges[pp][p]
		delete(t.edges[pp], p)
		delete(t.edges[p], pp)
		t.SetEdge(pp, b, be.e)
	}
	for _, n := range cycle {
		for _, m := range t.children[n] {
			if ContainsInt64(cycle, m) {
				continue
			}
			be := t.edges[n][m]
			delete(t.edges[n], m)
			delete(t.edges[m], n)
			t.SetEdge(b, m, be.e)
		}
	}

	// Set the heritage of the new blossom.
	// (Change root, parent, children info.)
	if pp != p {
		t.root[b] = t.root[p]
		t.parent[b] = pp
		t.children[pp] = []int64{b}
	} else {
		t.parent[b] = b
		t.root[b] = b
	}
	for _, n := range cycle {
		for _, m := range t.children[n] {
			if ContainsInt64(cycle, m) {
				continue
			}
			if _, ok := t.children[b]; ok {
				t.children[b] = append(t.children[b], m)
			} else {
				t.children[b] = []int64{m}
			}
			t.parent[m] = b
		}
	}
	if t.root[b] == b {
		t.ChangeRootFrom(b, b)
	}

	// Also, remove the tree structure of nodes in the cycle.
	for _, n := range cycle {
		delete(t.root, n)
		delete(t.parent, n)
		delete(t.children, n)
	}

	// Remove all matches in b.
	t.UnMatchBlossom(b)

	// Set new edges of b and match to its parent (if exists.)
	if t.root[b] != b {
		t.MatchEdgeBetween(b, t.parent[b])
	}
}

// CommonParent finds the common ansester of u, v and cycle of blossoms from p to m and to n.
// For example it returns [2 4 3 0 1] if the tree looks like:
//		p o  [2]
//	    /   \
// [1] o     o [4]
//     |     |
// [0] o     o [3]
//	   n     m

func (g *Tree) Cycle(n, m *Node) []*Node {
	hen := n.Heritage([]*Node{})
	hem := m.Heritage([]*Node{})
	var i1, i2 int
	for i1, nn := range hen {
		for i2, mm := range hem {
			if nn == mm {
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

func (g *Tree) ChangeRootFrom(b, r int64) {
	g.root[b] = r
	if _, ok := g.children[b]; !ok {
		return
	}
	for _, n := range g.children[b] {
		g.ChangeRootFrom(n, r)
	}
}
