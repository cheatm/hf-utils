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

func (n *Node[K, V]) String() string {
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
	// x: new node
	// p: x parent
	// g: p parent
	// u: x uncle or y sibling
	var x, p, g, tmp *Node[K, V]
	x = t.root
	for x != nil {
		p = x
		if key < x.key {
			x = x.left
		} else if x.key < key {
			x = x.right
		} else {
			x.value = value
			return
		}
	}

	// tree is empty
	if p == nil {
		t.root = &Node[K, V]{key: key, value: value, isBlack: BLACK}
		return
	}

	// x: new node
	x = &Node[K, V]{key: key, value: value, isBlack: RED, parent: p}
	if key < p.key {
		p.left = x
	} else {
		p.right = x
	}

	// t.InsertFix(c)

	for {
		// Loop invariant: node is red.
		if x == t.root {
			/*
			 * The inserted node is root. Either this is the
			 * first node, or we recursed at Case 1 below and
			 * are no longer violating 4).
			 */
			x.isBlack = BLACK
			break
		}

		if p.isBlack {
			/*
			 * If there is a black parent, we are done.
			 * Otherwise, take some corrective action as,
			 * per 4), we don't want a red root or two
			 * consecutive red nodes.
			 */
			break
		}

		g = p.parent
		tmp = g.right // tmp as u

		if p != tmp {
			// p = g.left
			if !IsBlack(tmp) {
				/*
				 * Case 1 - node's uncle is red (color flips).
				 *
				 *       G            g
				 *      / \          / \
				 *     p   u  -->   P   U
				 *    /            /
				 *   x            x
				 *
				 * However, since g's parent might be red, and
				 * 4) does not allow this, we need to recurse
				 * at g.
				 */
				p.isBlack = BLACK
				tmp.isBlack = BLACK
				g.isBlack = RED

				x = g
				p = g.parent

				continue
			}

			tmp = p.right
			if x == tmp {
				/*
				 * Case 2 - node's uncle is black and node is
				 * the parent's right child (left rotate at parent).
				 *
				 *      G             G
				 *     / \           / \
				 *    p   U  -->    x   U
				 *     \           /
				 *      x         p
				 *     /           \
				 *    L             L
				 * This still leaves us in violation of 4), the
				 * continuation into Case 3 will fix that.
				 */

				p.AddRight(x.left)
				x.AddLeft(p)
				g.AddLeft(x)
				p = x
				tmp = p.right
			}

			/*
			 * Case 3 - node's uncle is black and node is
			 * the parent's left child (right rotate at gparent).
			 *
			 *        G           P
			 *       / \         / \
			 *      p   U  -->  x   g
			 *     / \             / \
			 *    x  (r)          R   U
			 */

			if tmp != nil {
				g.AddLeft(tmp)
				tmp.isBlack = BLACK
			} else {
				g.left = nil
			}
			g.isBlack = RED

			if g == t.root {
				t.root = p
				p.parent = nil
			} else {
				if g == g.parent.left {
					g.parent.AddLeft(p)
				} else {
					g.parent.AddRight(p)
				}
			}

			p.AddRight(g)
			p.isBlack = BLACK
			break

		} else {
			// p = g.right
			tmp = g.left
			if !IsBlack(tmp) {
				/*
				 * Case 1 - node's uncle is red (color flips).
				 *
				 *       G            g
				 *      / \          / \
				 *     u   p  -->   U   P
				 *        /            /
				 *       x            x
				 *
				 * However, since g's parent might be red, and
				 * 4) does not allow this, we need to recurse
				 * at g.
				 */
				p.isBlack = BLACK
				tmp.isBlack = BLACK
				g.isBlack = RED

				x = g
				p = g.parent

				continue
			}

			tmp = p.left
			if x == tmp {
				/*
				 * Case 2 - node's uncle is black and node is
				 * the parent's left child (right rotate at parent).
				 *
				 *      G             G
				 *     / \           / \
				 *    U   p  -->    U   x
				 *       /               \
				 *      x                 p
				 *       \               /
				 *        R             R
				 * This still leaves us in violation of 4), the
				 * continuation into Case 3 will fix that.
				 */

				p.AddLeft(x.right)
				x.AddRight(p)
				g.AddRight(x)
				p = x
				tmp = p.left
			}

			/*
			 * Case 3 - node's uncle is black and node is
			 * the parent's left child (right rotate at gparent).
			 *
			 *        G           P
			 *       / \         / \
			 *      U   p  -->  g   x
			 *         / \     / \
			 *       (l)  x   U   L
			 */

			if tmp != nil {
				g.AddRight(tmp)
				tmp.isBlack = BLACK
			} else {
				g.right = nil
			}
			g.isBlack = RED

			if g == t.root {
				t.root = p
				p.parent = nil
			} else {
				if g == g.parent.left {
					g.parent.AddLeft(p)
				} else {
					g.parent.AddRight(p)
				}
			}

			p.AddLeft(g)
			p.isBlack = BLACK
			break
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
	// set x as node's prev
	var x *Node[K, V]
	if node.left != nil {
		x = node.left
		for x.right != nil {
			x = x.right
		}

		// replace node with x
		node.key = x.key
		node.value = x.value
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
