package rbtree

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type Color bool

const (
	RED   Color = false
	BLACK Color = true
)

type Node[K constraints.Ordered, V interface{}] struct {
	left    *Node[K, V]
	right   *Node[K, V]
	parent  *Node[K, V]
	isBlack Color
	key     K
	value   V
}

func (node *Node[K, V]) AddLeft(child *Node[K, V]) {
	node.left = child
	if child != nil {
		child.parent = node
	}
}

func (node *Node[K, V]) AddRight(child *Node[K, V]) {
	node.right = child
	if child != nil {
		child.parent = node
	}
}

func (node *Node[K, V]) Sibling() *Node[K, V] {
	if node == node.parent.left {
		return node.parent.right
	} else {
		return node.parent.left
	}
}

func (n *Node[K, V]) GoString() string {
	if n.isBlack {
		return fmt.Sprintf("(%v:%v)", n.key, n.value)
	} else {
		return fmt.Sprintf("[%v:%v]", n.key, n.value)
	}
}

func (n *Node[K, V]) KeyString() string {
	if n.isBlack {
		return fmt.Sprintf("(%v)", n.key)
	} else {
		return fmt.Sprintf("[%v]", n.key)
	}
}

func (n *Node[K, V]) Pair() (K, V) {
	return n.key, n.value
}

type RBTree[K constraints.Ordered, V interface{}] struct {
	root *Node[K, V]
}

func (t *RBTree[K, V]) Insert(key K, value V) {
	var y *Node[K, V] = nil
	x := t.root
	for x != nil {
		y = x
		if key < x.key {
			x = x.left
		} else if x.key < key {
			x = x.right
		} else {
			x.value = value
			return
		}
	}

	if y == nil {
		t.root = &Node[K, V]{key: key, value: value, isBlack: BLACK}
	} else {
		c := &Node[K, V]{key: key, value: value, isBlack: RED, parent: y}
		if key < y.key {
			y.left = c
		} else {
			y.right = c
		}

		t.InsertFix(c)

	}

}

func (t *RBTree[K, V]) InsertFix(x *Node[K, V]) {
	p := x.parent
	if p.isBlack {
		return
	}
	g := p.parent
	var u *Node[K, V]
	pl := false
	if g.left == p {
		u = g.right
		pl = true
	} else {
		u = g.left
	}

	if u != nil && !u.isBlack {
		// Uncle, Parent: [RED]
		// GrandParent: (BLACK)
		//
		//       (g)       ->[g]
		//       / \         / \
		//     [p] [u]  => (p) (u)
		//     /           /
		// ->[x]         [x]

		u.isBlack = BLACK
		p.isBlack = BLACK
		if g != t.root {
			g.isBlack = RED
			t.InsertFix(g)
		}

		return
	} else {
		// Parent: [RED]
		// Grand: (BLACK)
		// Uncle: (BLACK) or (nil)

		if p.left == x {
			if pl {
				// LL:
				//     (g)            (p)
				//     / \            / \
				//   [p] (u)   =>   [x] [g]
				//   / \                / \
				// [x] (o)            (o) (u)

				// [P] -> (p) and raise (p)
				t.raiseNode(p, g)

				// (o) -> (g).left
				g.AddLeft(p.right)

				// (g) -> [g] -> (p).right
				g.isBlack = RED
				p.AddRight(g)

			} else {
				// RL
				//   (g)              (x)
				//   / \            /     \
				// (u) [p]   =>   [g]     [p]
				//     /          / \     /
				//   [x]        (u) (l) (r)
				//   / \
				// (l) (r)

				// [x] -> (x) and raise (x)
				t.raiseNode(x, g)
				// (r) -> [p].left
				p.AddLeft(x.right)

				// (l) -> [g].right
				g.AddRight(x.left)

				// (g) -> [g] -> (x).left
				g.isBlack = RED
				x.AddLeft(g)

				// [p] -> (x).right
				x.AddRight(p)

			}

		} else {
			if pl {
				// LR:
				//   (g)              (x)
				//   / \            /     \
				// [p] (u)   =>   [p]     [g]
				//   \              \     / \
				//   [x]            (l) (r) (u)
				//   / \
				// (l) (r)

				// [x] -> (x) and raise (x)
				t.raiseNode(x, g)

				// (l) -> [p].right
				p.AddRight(x.left)

				// (r) -> [g].left
				g.AddLeft(x.right)

				// (g) -> [g] -> (x).right
				g.isBlack = RED
				x.AddRight(g)

				// [p] -> (x).left
				x.AddLeft(p)

			} else {
				// RR
				//   (g)            (p)
				//   / \            / \
				// (u) [p]   =>   [g] [x]
				//     / \        / \
				//   (o) [x]    (u) (o)

				// [P] -> (p) and raise (p)
				t.raiseNode(p, g)

				// (o) -> (g).right
				g.AddRight(p.left)
				// (g) -> [g] -> (p).left
				g.isBlack = RED
				p.AddLeft(g)

			}

		}

	}

}

// [x] -> (x) and replace (g)'s position
func (t *RBTree[K, V]) raiseNode(x, g *Node[K, V]) {
	x.isBlack = BLACK
	if t.root != g {
		// g is not root
		if g.parent.left == g {
			g.parent.AddLeft(x)
		} else {
			g.parent.AddRight(x)
		}
	} else {
		// g is root
		t.root = x
		x.parent = nil
	}
}

func (t *RBTree[K, V]) Delete(key K) bool {
	node := t.GetNode(key)
	if node == nil {
		return false
	}
	// set x as node's next
	var x *Node[K, V]
	if node.right != nil {
		x = node.right
		for x.left != nil {
			x = x.left
		}
		x.key = node.key
		x.value = node.value
	} else {
		x = node
	}

	//    x
	//   / \
	// nil  c

	var c *Node[K, V]
	if x.left == nil {
		c = x.right
	} else {
		c = x.left
	}

	// replace x with c
	xl := false
	if x == t.root {
		t.root = c
		if c != nil {
			c.parent = nil
		} else {
			return true
		}

	} else {
		if x == x.parent.left {
			x.parent.AddLeft(c)
			xl = true
		} else {
			x.parent.AddRight(c)
		}
	}

	if x.isBlack {
		if IsBlack(c) {
			t.FixReplacedX(c, x.parent, xl)
		} else {
			c.isBlack = BLACK
		}

	}

	return true
}

func (t *RBTree[K, V]) FixReplacedX(x *Node[K, V], p *Node[K, V], xl bool) {
	if p == nil {
		return
	}
	if p.isBlack {
		if xl {
			s := p.right
			if IsBlack(s) {
				t.FixBBB(p, s)
			} else {
				t.FixXBR(p, s)
			}
		} else {
			s := p.left
			if IsBlack(s) {
				t.FixBBB(p, s)
			} else {
				t.FixRBX(p, s)
			}
		}
	} else {
		if xl {
			t.FixXRB(p, p.right)
		} else {
			t.FixBRX(p, p.left)
		}
	}
}

func IsBlack[K constraints.Ordered, V interface{}](x *Node[K, V]) Color {
	if x == nil {
		return BLACK
	}
	return x.isBlack
}

func (t *RBTree[K, V]) FixBRX(p, s *Node[K, V]) {
	// case 4
	//   [p]
	//   / \
	// (s) (x)
START:
	if IsBlack(s.left) && IsBlack(s.right) {
		//   (s)
		//   / \
		// (l) (r)
		s.isBlack = RED
		p.isBlack = BLACK
	} else if !IsBlack(s.left) {
		//     [p]
		//     /
		//   (s)
		//   / \
		//  [l] r
		if p == t.root {
			t.root = s
			s.parent = nil
		} else {
			if p == p.parent.right {
				p.parent.right = s
			} else {
				p.parent.left = s
			}
			s.parent = p.parent
		}
		s.isBlack = p.isBlack

		p.AddLeft(s.right)
		p.isBlack = BLACK
		s.AddRight(s.left)

	} else {
		//   (s)
		//   / \
		// (l) [r]
		sr := s.right
		p.left = s
		sr.parent = p
		sr.isBlack = BLACK

		s.right = sr.right
		s.right.parent = s

		sr.left = s
		s.parent = sr
		s.isBlack = RED
		s = sr
		//       [p]
		//       /
		//     (r)
		//     /
		//   [s]
		//   /
		// (l)
		goto START
	}
}

func (t *RBTree[K, V]) FixXRB(p, s *Node[K, V]) {
	// case 4
	//   [p]
	//   / \
	// (x) (s)
START:
	if IsBlack(s.left) && IsBlack(s.right) {
		//   (s)
		//   / \
		// (l) (r)
		s.isBlack = RED
		p.isBlack = BLACK
	} else if !IsBlack(s.right) {
		//  [p]
		//    \
		//   (s)
		//   / \
		//  l  [r]
		if p == t.root {
			t.root = s
			s.parent = nil
		} else {
			if p == p.parent.left {
				p.parent.left = s
			} else {
				p.parent.right = s
			}
			s.parent = p.parent
		}
		s.isBlack = p.isBlack

		p.AddRight(s.left)
		p.isBlack = BLACK
		s.AddLeft(p)

	} else {
		//   (s)
		//   / \
		// [l] (r)
		sl := s.left
		p.right = sl
		sl.parent = p
		sl.isBlack = BLACK

		s.left = sl.right
		s.left.parent = s

		sl.right = s
		s.parent = sl
		s.isBlack = RED
		s = sl
		//  [p]
		//    \
		//   (l)
		//     \
		//     [s]
		//      \
		//      (r)
		goto START
	}

}

func (t *RBTree[K, V]) FixBBB(p, s *Node[K, V]) {
	// case 3

	//   (p)         (p)  -> set p as x
	//   / \          \
	// (x) (s) =>    [s]
	//     / \
	//   (l) (r)

	//     (p)
	//     / \
	//   (s) (x)
	//   / \
	// (l) (r)

	s.isBlack = RED
	if p.parent == nil {
		t.FixReplacedX(p, p.parent, false)
	} else {
		t.FixReplacedX(p, p.parent, p == p.parent.left)
	}
}

func (t *RBTree[K, V]) FixXBR(p, s *Node[K, V]) {
	// case 2
	//   (p)           (s)
	//   / \           / \
	// (x) [s]   =>  [p] (r)
	//     / \       / \
	//   (l) (r)   (x) (l)

	// if p == t.root {
	// 	t.root = s
	// 	s.parent = nil
	// } else {
	// 	if p == p.parent.left {
	// 		p.parent.AddLeft(s)
	// 	} else {
	// 		p.parent.AddRight(s)
	// 	}
	// }
	// s.isBlack = BLACK
	t.XReplaceN(s, p)

	p.AddRight(s.left)
	s.AddLeft(p)
	p.isBlack = RED
	t.FixXRB(p, p.right)
}

func (t *RBTree[K, V]) FixRBX(p, s *Node[K, V]) {
	// case 2
	//     (p)            (s)
	//     / \            / \
	//   [s] (x)   =>   (l) [p]
	//   / \                / \
	// (l) (r)   	  	  (r) (x)

	// if p == t.root {
	// 	t.root = s
	// 	s.parent = nil
	// } else {
	// 	if p == p.parent.right {
	// 		p.parent.AddRight(s)
	// 	} else {
	// 		p.parent.AddLeft(s)
	// 	}
	// }
	// s.isBlack = BLACK
	t.XReplaceN(s, p)

	p.AddLeft(s.right)
	s.AddRight(p)
	p.isBlack = RED
	t.FixBRX(p, p.left)
}

func (t *RBTree[K, V]) XReplaceN(x, n *Node[K, V]) {
	if n != t.root {
		if n == n.parent.left {
			n.parent.AddRight(x)
		} else {
			n.parent.AddLeft(x)
		}
		x.isBlack = n.isBlack
	} else {
		t.root = x
		x.parent = nil
		x.isBlack = BLACK
	}
}

func (t *RBTree[K, V]) Get(key K) (value V, ok bool) {
	node := t.root
	for node != nil {
		if key < node.key {
			node = node.left
		} else if node.key < key {
			node = node.right
		} else {
			value = node.value
			ok = true
			break
		}
	}
	return
}

func (t *RBTree[K, V]) GetNode(key K) *Node[K, V] {
	node := t.root
	for node != nil {
		if key < node.key {
			node = node.left
		} else if node.key < key {
			node = node.right
		} else {
			break
		}
	}
	return node
}
