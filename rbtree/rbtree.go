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

func (n *Node[K, V]) GoString() string {
	if n.isBlack {
		return fmt.Sprintf("B{%v:%v}", n.key, n.value)
	} else {
		return fmt.Sprintf("R{%v:%v}", n.key, n.value)
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
	if g.parent != nil {
		// g is not root
		if g.parent.left == g {
			g.parent.AddLeft(x)
		} else {
			g.parent.AddRight(x)
		}
	} else {
		// g is root
		t.root = x
	}
}

func (t *RBTree[K, V]) Delete(key K) bool {

	return false
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
