package mwpm

import "math"

type Node struct {
	label           int
	parent          *Node
	blossom         *Node // blossom that contains this node (nil if the node is not in a blossom )
	cycle, children []*Node
	dval            float64
}

const Eps = 1e-6

func NewNode() *Node {
	return &Node{
		label:  1,
		parent: nil,
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

func (n *Node) Anscesters() []*Node {
	if n.parent == nil {
		return []*Node{n}
	}
	return append([]*Node{n}, n.parent.Anscesters()...)
}

func (n *Node) Descendants() []*Node {
	if len(n.children) == 0 {
		return []*Node{n}
	}
	nodes := []*Node{n}
	for _, c := range n.children {
		nodes = append(nodes, c.Descendants()...)
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
