package mwpm

import "math"

type Node struct {
	label    int
	parent   *Node   // directs to non-blossom node
	children []*Node // directs to non-blossom node
	blossom  *Node   // immediate blossom that contains this node (nil if the node is the outermost blossom)
	// blossom *Blossom
	cycle [][3]*Node // cycle of nodes (blossom, start, end)
	dval  float64
	temp  int64
}

const Eps = 1e-6

func NewNode() *Node {
	return &Node{
		label:   1,
		parent:  nil,
		blossom: nil,
	}
}

// returns the outermost blossom
func (n *Node) Blossom() *Node {
	if n.blossom == nil {
		return n
	}
	return n.blossom.Blossom()
}

func (n *Node) Root() *Node {
	if n.parent == nil {
		return n
	}
	return n.parent.Root()
}

func (n *Node) anscesters() []*Node {
	if n.parent == nil {
		return []*Node{n}
	}
	return append([]*Node{n}, n.parent.anscesters()...)
}

// return ALL node (not blossom) descendants from the blossom containing n
func (n *Node) descendants() []*Node {
	nb := n.Blossom()
	if len(nb.children) == 0 {
		if nb == n {
			return []*Node{n}
		}
		return n.all()
	}
	nodes := []*Node{n}
	for _, c := range n.children {
		nodes = append(nodes, c.descendants()...)
	}
	return nodes
}

func (n *Node) all() []*Node {
	if n.blossom == nil {
		return []*Node{n}
	}
	nodes := []*Node{}
	for _, c := range n.cycle {
		nodes = append(nodes, c[0].all()...)
	}
	return nodes
}

func (n *Node) IsDvalZero() bool {
	return math.Abs(n.dval) < Eps
}

func (n *Node) IsBlossom() bool {
	return len(n.cycle) > 1
}

func (n *Node) RemoveChild(m *Node) {
	for i, c := range n.children {
		if c == m {
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
}
