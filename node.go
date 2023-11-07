package mwpm

type Node struct {
	label           int
	parent, blossom *Node
	cycle           []*Node
	children        map[*Node]struct{}
	dval            float64
}

func NewNode() *Node {
	return &Node{
		label:   1,
		parent:  nil,
		blossom: nil,
	}
}

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

func (n *Node) Heritage(her []*Node) []*Node {
	if n.parent == nil {
		return append(her, n)
	}
	return n.parent.Heritage(append(her, n))
}
