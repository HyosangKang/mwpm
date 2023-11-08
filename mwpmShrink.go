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
	nn := NewNode()
	var cycle []*Node = t.Cycle(n, m)  // cycle, the first is the common parent
	var comm *Node = cycle[0]          // common parent of n, m
	nn.parent = comm.parent            // grandparent is now new parent
	delete(comm.parent.children, comm) // remove the child (the parent of n, m) from grandparent
	comm.children[nn] = struct{}{}
	nn.cycle = cycle
	for _, n := range cycle {
		n.label = 0
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
	var nn, mm *Node
	for i1, nn = range hen {
		for i2, mm = range hem {
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
