package mwpm

// Expand removes b and add nodes in b to the tree.
// The decision which nodes belong to tree depends on positions of nodes in the cycle.
// For example, if we expand a blossom [5] consists of [1,2,3], we obtain
// [4] o                       [4] o
//
//	|                           |
//
// [3] o                       [3] o +-+ o [2]
//
//	| \          ------->            /
//
// [1] o - o [2]               [1] o --
//
//	|                           I
//
// [0] o                       [0] o
// Nodes that are not added to the tree is matched pairwise.
// Since b is a negative blossom (>3 nodes),
// there are always a parent and a child of negative blossom
func (g *Tree) Expand(b int64) {
	p := g.parent[b]
	c := g.children[b][0]

	// Find nodes bp, bc in b which are matched to the parent and child of b respectively.
	be := g.edges[p][b]
	n := be.e.From().ID()
	if !ContainsInt64(g.nodes[b], n) {
		n = be.e.To().ID()
	}
	var bp int64
	for _, f := range g.cycle[b] {
		if ContainsInt64(g.nodes[f], n) {
			bp = f
			break
		}
	}
	be = g.edges[c][b]
	n = be.e.From().ID()
	if !ContainsInt64(g.nodes[b], n) {
		n = be.e.To().ID()
	}
	var bc int64
	for _, f := range g.cycle[b] {
		if ContainsInt64(g.nodes[f], n) {
			bc = f
			break
		}
	}

	// Reorder the cycle of b as [bp] to [bc]
	var i int
	cycle := g.cycle[b]
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
	children := g.children[p]
	for i = 0; i < len(children); i++ {
		if children[i] == b {
			break
		}
	}
	children = append(children[:i], children[i+1:]...)
	g.children[p] = append(children, bp)
	for i = 0; i < ic; i++ {
		g.children[cycle[i]] = []int64{cycle[i+1]}
	}
	g.children[cycle[ic]] = []int64{c}

	// Update parent info.
	g.parent[c] = bc
	for i = ic; i > 0; i-- {
		g.parent[cycle[i]] = cycle[i-1]
	}
	g.parent[cycle[0]] = p
	g.parent[c] = bc

	// Update root info.
	for i = 0; i <= ic; i++ {
		g.root[cycle[i]] = g.root[p]
	}

	// Update label info.
	for i = 0; i <= ic; i++ {
		if i%2 == 0 {
			g.label[cycle[i]] = -1
		} else {
			g.label[cycle[i]] = 1
		}
	}
	for i = ic + 1; i < len(cycle); i++ {
		g.LabelAsZero(cycle[i])
	}

	// Set edge between p and bp. Also set edge between c and bc, and match
	be = g.edges[b][p]
	g.SetEdge(p, bp, be.e)
	g.RemoveEdge(p, b)
	be = g.edges[b][c]
	g.SetEdge(c, bc, be.e)
	g.RemoveEdge(b, c)

	// Remove edge between rest of cycle
	for i = ic + 1; i < len(cycle); i += 2 {
		g.RemoveEdge(cycle[i], cycle[i-1])
	}
	g.RemoveEdge(cycle[len(cycle)-1], cycle[0])

	// Apply matching. Match new nodes in the tree as bp -- bc -- c.
	// Also, match the free nodes in the cycle.
	for i = 1; i < ic; i += 2 {
		g.MatchEdgeBetween(cycle[i], cycle[i-1])
	}
	g.MatchEdgeBetween(c, bc)
	for i = ic + 2; i < len(cycle); i += 2 {
		g.MatchEdgeBetween(cycle[i], cycle[i-1])
	}

	// Free cycle nodes from the blossom
	for i = 0; i < len(cycle); i++ {
		delete(g.blossom, cycle[i])
	}

	// Remove the blossom permanently.
	// Do not remove label because this count the total number of nodes used.
	g.label[b] = -2
	delete(g.nodes, b)
	delete(g.cycle, b)
	delete(g.root, b)
	delete(g.children, b)
	delete(g.dval, b)
	delete(g.parent, b)
	delete(g.edges, b)
}

// RemoveEdge deletes the blossom edge between blossom u and v.
func (g *Tree) RemoveEdge(u, v int64) {
	delete(g.edges[u], v)
	delete(g.edges[v], u)
}
