package mwpm

type Node struct {
	label    int
	parent   *Node      // directs to non-blossom node
	children []*Node    // directs to non-blossom node
	blossom  *Node      // immediate blossom that contains this node (nil if the node is the outermost blossom)
	cycle    [][2]*Node // cyclic pair of nodes (start, end)
	dval     float64
	temp     int64
}

// returns the outermost blossom
func (n *Node) Blossom() *Node {
	if n.blossom == nil {
		return n
	}
	return n.blossom.Blossom()
}

func (n *Node) Root() *Node {
	n = n.Blossom()
	for n.parent != nil {
		n = n.parent.Blossom()
	}
	return n
}

func (n *Node) Anscesters() []*Node {
	n = n.Blossom()
	chain := []*Node{n}
	for n.parent != nil {
		chain = append(chain, n.parent.Blossom())
		n = n.parent.Blossom()
	}
	return chain
}

func (n *Node) Anscestary() [][2]*Node {
	n = n.Blossom()
	cycle := [][2]*Node{}
	for n.parent != nil {
		for _, u := range n.parent.Blossom().children {
			if u.Blossom() == n {
				cycle = append(cycle, [2]*Node{u, n.parent})
				break
			}
		}
		n = n.parent.Blossom()
	}
	return cycle
}

// return ALL child blossom nodes from n
func (n *Node) Descendents() []*Node {
	n = n.Blossom()
	if len(n.children) == 0 {
		return []*Node{n}
	}
	nodes := []*Node{n}
	for _, c := range n.children {
		nodes = append(nodes, c.Blossom().Descendents()...)
	}
	return nodes
}

// returns All nodes (not blossom) in the blossom n
func (n *Node) All() []*Node {
	if len(n.cycle) == 0 {
		return []*Node{n}
	}
	nodes := []*Node{}
	for _, c := range n.cycle {
		nodes = append(nodes, c[0].BlossomWithin(n).All()...)
	}
	return nodes
}

func (n *Node) BlossomWithin(b *Node) *Node {
	for n.blossom != b {
		n = n.blossom
	}
	if n.blossom == nil {
		panic("invalid search for blossom within")
	}
	return n
}

func (n *Node) RemoveParent() {
	for n.blossom != nil {
		n.parent = nil
		n = n.blossom
	}
}

func (n *Node) PopChild(m *Node) *Node {
	for l, u := range n.children {
		ub := u.Blossom()
		if ub == m {
			n.children = append(n.children[:l], n.children[l+1:]...)
			return u
		}
	}
	return nil
}

func (n *Node) SetAlone() {
	n.parent = nil
	n.children = []*Node{}
}

// set the blossom (or nodes) as a free node
func (b *Node) SetFree() {
	// fmt.Printf("set free %d (%v)\n", b.temp, b)
	b.label = 0
	b.SetAlone()
	for _, c := range b.cycle {
		c[0].BlossomWithin(b).SetFree()
	}
}

func (n *Node) Update(delta float64) {
	n.dval += float64(n.label) * delta
	for _, c := range n.cycle {
		cb := c[0].BlossomWithin(n)
		cb.Update(delta)
	}
}
