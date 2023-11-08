package mwpm

// Augment increases the matching. It is desribed in the following pictorial way:
// (+) o        o (+)                o     o
//	    \     /   \                  I     I
//   (-) o   o (-) o (-)   ----->    o     o   o
//       I   I     I                           I
//   (+) o - o (+) o (+)             o +-+ o   o
//       u   v

func (t *Tree) Augment(n, m *Node) {
	t.match[n], t.match[m] = m, n
	for _, l := range [2]*Node{n, m} {
		for u := l; u.parent != nil; u = u.parent {
			if u.label < 0 {
				t.match[u], t.match[u.parent] = u.parent, u
			}
		}
		for _, v := range l.Root().Descendants() {
			t.Free(v)
		}
	}
}
