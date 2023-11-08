package mwpm

// Expand removes b and add nodes in b to the tree.
// The decision which nodes belong to tree depends on positions of nodes in the cycle.
// For example, if we expand a blossom [5] consists of [1,2,3], we obtain
//                   [4] o                       [4] o
//                       |                           |
//  [3] o --- o [2]    n o                       [3] o +-+ o [2]
//      |  n  /          |            ------->            /
//  [1] o --             |                       [1] o --
//                       |                           I
//                   [0] o                       [0] o
// Nodes that are not added to the tree is matched pairwise.
// Since b is a negative blossom (>3 nodes),
// there are always a parent and a child of negative blossom

func (t *Tree) Expand(n *Node) {
	/* change the children of [4] from b to [3] */
	delete(n.parent.children, n)
	n.parent.children[n.cycle[0]] = struct{}{}
	/* traverse the cycle from [3] until [1] */
	var i int
	lab := -1
	for i = 0; len(n.cycle[i].children) == 0; i++ {
		if i == 0 {
			n.cycle[i].parent = n.parent
		} else {
			n.cycle[i].parent = n.cycle[i-1]
		}
		n.cycle[i].label = lab
		n.cycle[i].children[n.cycle[i+1]] = struct{}{}
		lab *= -1
	}
	n.cycle[i].children[0].parent = n.cycle[i]
	delete(t.match, n)
	delete(t.match, n.cycle[i].children[0])
	for j := i + 1; j < len(n.cycle); j += 2 {
		t.match[n.cycle[j]] = n.cycle[j+1]
		for k := 0; k < 2; k++ {
			n.cycle[j+k].parent = nil
			n.cycle[j+k].label = 0
		}
	}
	// Find nodes bp, bc in b which are matched to the parent and child of b respectively.
	be := t.edges[p][b]
	n := be.e.From().ID()
	if !ContainsInt64(t.nodes[b], n) {
		n = be.e.To().ID()
	}
	var bp int64
	for _, f := range t.cycle[b] {
		if ContainsInt64(t.nodes[f], n) {
			bp = f
			break
		}
	}
	be = t.edges[c][b]
	n = be.e.From().ID()
	if !ContainsInt64(t.nodes[b], n) {
		n = be.e.To().ID()
	}
	var bc int64
	for _, f := range t.cycle[b] {
		if ContainsInt64(t.nodes[f], n) {
			bc = f
			break
		}
	}

	// Reorder the cycle of b as [bp] to [bc]
	var i int
	cycle := t.cycle[b]
	for i = 0; i < len(cycle); i++ {
		if cycle[i] == bp {
			break
		}
	}
	cycle = append(cycle[i:], cycle[:i]...)

	// If the index of [bc] in the cycle is odd, reverse the order.
	// ic is the place of bc in the cycle.
	var ic int
	for ic = 0; ic < len(cycle); ic++ {
		if cycle[ic] == bc {
			break
		}
	}
	if ic%2 == 1 {
		rev := []int64{cycle[0]}
		for j := 1; j < len(cycle); j++ {
			rev = append(rev, cycle[len(cycle)-j])
		}
		cycle = rev
		ic = len(cycle) - ic
	}

	// Remove b from children of p and add bp as new children.
	// Assign children to nodes in cycle in order.
	children := t.children[p]
	for i = 0; i < len(children); i++ {
		if children[i] == b {
			break
		}
	}
	children = append(children[:i], children[i+1:]...)
	t.children[p] = append(children, bp)
	for i = 0; i < ic; i++ {
		t.children[cycle[i]] = []int64{cycle[i+1]}
	}
	t.children[cycle[ic]] = []int64{c}

	// Update parent info.
	t.parent[c] = bc
	for i = ic; i > 0; i-- {
		t.parent[cycle[i]] = cycle[i-1]
	}
	t.parent[cycle[0]] = p
	t.parent[c] = bc

	// Update root info.
	for i = 0; i <= ic; i++ {
		t.root[cycle[i]] = t.root[p]
	}

	// Update label info.
	for i = 0; i <= ic; i++ {
		if i%2 == 0 {
			t.label[cycle[i]] = -1
		} else {
			t.label[cycle[i]] = 1
		}
	}
	for i = ic + 1; i < len(cycle); i++ {
		t.LabelAsZero(cycle[i])
	}

	// Set edge between p and bp. Also set edge between c and bc, and match
	be = t.edges[b][p]
	t.SetEdge(p, bp, be.e)
	t.RemoveEdge(p, b)
	be = t.edges[b][c]
	t.SetEdge(c, bc, be.e)
	t.RemoveEdge(b, c)

	// Remove edge between rest of cycle
	for i = ic + 1; i < len(cycle); i += 2 {
		t.RemoveEdge(cycle[i], cycle[i-1])
	}
	t.RemoveEdge(cycle[len(cycle)-1], cycle[0])

	// Apply matching. Match new nodes in the tree as bp -- bc -- c.
	// Also, match the free nodes in the cycle.
	for i = 1; i < ic; i += 2 {
		t.MatchEdgeBetween(cycle[i], cycle[i-1])
	}
	t.MatchEdgeBetween(c, bc)
	for i = ic + 2; i < len(cycle); i += 2 {
		t.MatchEdgeBetween(cycle[i], cycle[i-1])
	}

	// Free cycle nodes from the blossom
	for i = 0; i < len(cycle); i++ {
		delete(t.blossom, cycle[i])
	}

	// Remove the blossom permanently.
	// Do not remove label because this count the total number of nodes used.
	t.label[b] = -2
	delete(t.nodes, b)
	delete(t.cycle, b)
	delete(t.root, b)
	delete(t.children, b)
	delete(t.dval, b)
	delete(t.parent, b)
	delete(t.edges, b)
}

// RemoveEdge deletes the blossom edge between blossom u and v.
func (g *Tree) RemoveEdge(u, v int64) {
	delete(g.edges[u], v)
	delete(g.edges[v], u)
}
