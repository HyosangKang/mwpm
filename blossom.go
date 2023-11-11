package mwpm

import (
	"fmt"
	"math"
)

type Blossom struct {
	label    int
	parent   *Blossom   // directs to non-blossom node
	children []*Blossom // directs to non-blossom node
	blossom  *Blossom   // immediate blossom that contains this node (nil if the node is the outermost blossom)
	// blossom *Blossom
	cycle [][2]*Blossom // cyclic pair of nodes (start, end)
	// chain []*Node    // all nodes in the blossom in the same cyclic order as cycle ()
	dval float64
	temp int64
}

const Eps = 1e-6

func NewNode() *Blossom {
	return &Blossom{
		label:    1,
		children: []*Blossom{},
		cycle:    [][2]*Blossom{},
	}
}

// returns the outermost blossom
func (n *Blossom) Blossom() *Blossom {
	if n.blossom == nil {
		return n
	}
	return n.blossom.Blossom()
}

func (n *Blossom) Root() *Blossom {
	n = n.Blossom()
	for n.parent != nil {
		n = n.parent.Blossom()
	}
	return n
}

func (n *Blossom) anscesters() []*Blossom {
	n = n.Blossom()
	chain := []*Blossom{n}
	for n.parent != nil {
		chain = append(chain, n.parent.Blossom())
		n = n.parent.Blossom()
	}
	return chain
}

func (n *Blossom) anscestary() [][2]*Blossom {
	n = n.Blossom()
	cycle := [][2]*Blossom{}
	for n.parent != nil {
		for _, u := range n.parent.Blossom().children {
			if u.Blossom() == n {
				cycle = append(cycle, [2]*Blossom{u, n.parent})
				break
			}
		}
		n = n.parent.Blossom()
	}
	return cycle
}

// return ALL child blossom nodes from n
func (n *Blossom) descendents() []*Blossom {
	n = n.Blossom()
	if len(n.children) == 0 {
		return []*Blossom{n}
	}
	nodes := []*Blossom{n}
	for _, c := range n.children {
		nodes = append(nodes, c.Blossom().descendents()...)
	}
	return nodes
}

// returns all nodes (not blossom) in the blossom n
func (n *Blossom) all() []*Blossom {
	if len(n.cycle) == 0 {
		return []*Blossom{n}
	}
	nodes := []*Blossom{}
	for _, c := range n.cycle {
		nodes = append(nodes, c[0].BlossomWithin(n).all()...)
	}
	return nodes
}

func (n *Blossom) IsDvalZero() bool {
	return math.Abs(n.dval) < Eps
}

func (n *Blossom) IsBlossom() bool {
	return len(n.cycle) > 1
}

func (n *Blossom) RemoveChild(m *Blossom) {
	for i, c := range n.children {
		if c == m {
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
}

func (n *Blossom) BlossomWithin(b *Blossom) *Blossom {
	for n.blossom != b {
		n = n.blossom
	}
	if n.blossom == nil {
		panic("invalid search for blossom within")
	}
	return n
}

func (n *Blossom) AllBlossoms() []*Blossom {
	blossoms := []*Blossom{n}
	for n.blossom != nil {
		blossoms = append(blossoms, n.blossom)
		n = n.blossom
	}
	return blossoms
}

func (n *Blossom) RemoveParent() {
	for n.blossom != nil {
		n.parent = nil
		n = n.blossom
	}
}

func (n *Blossom) show() {
	fmt.Printf("Node id %d label %d\n", n.temp, n.label)
}

func (n *Blossom) popChild(m *Blossom) *Blossom {
	for l, u := range n.children {
		ub := u.Blossom()
		if ub == m {
			n.children = append(n.children[:l], n.children[l+1:]...)
			return u
		}
	}
	return nil
}

func (n *Blossom) SetAlone() {
	n.parent = nil
	n.children = []*Blossom{}
}

func (n *Blossom) update(delta float64) {
	n.dval += float64(n.label) * delta
	for _, c := range n.cycle {
		cb := c[0].BlossomWithin(n)
		cb.update(delta)
	}
}
