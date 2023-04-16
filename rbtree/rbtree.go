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

func (t *RBTree[K, V]) fixDelete(n, p *Node[K, V]) {
	// rebalance on n after delete
	//
	// s: n sibling
	// p: n parent
	var s, tmp1, tmp2 *Node[K, V]
	for {
		s = p.right
		if n != s {
			// n = p.left
			if !IsBlack(s) {
				/*
				 * Case 1 - left rotate at parent
				 *
				 *     P               S
				 *    / \             / \
				 *   N   s    -->    p   Sr
				 *      / \         / \
				 *     Sl  Sr      N   Sl
				 */
				tmp1 = s.left
				p.AddRight(tmp1)
				t.XReplaceN(s, p)
				s.isBlack = BLACK
				s.AddLeft(p)
				p.isBlack = RED
				s = tmp1
			}
			tmp1 = s.right
			if IsBlack(tmp1) {
				tmp2 = s.left
				if IsBlack(tmp2) {
					/*
					 * Case 2 - sibling color flip
					 * (p could be either color here)
					 *
					 *    (p)           (p)
					 *    / \           / \
					 *   N   S    -->  N   s
					 *      / \           / \
					 *     Sl  Sr        Sl  Sr
					 *
					 * This leaves us violating 5) which
					 * can be fixed by flipping p to black
					 * if it was red, or by recursing at p.
					 * p is red when coming from Case 1.
					 */
					s.isBlack = RED
					if !p.isBlack {
						p.isBlack = BLACK
					} else {
						if p != t.root {
							n = p
							p = n.parent
							continue
						}
					}
					break
				}
				/*
				 * Case 3 - right rotate at sibling
				 * (p could be either color here)
				 *
				 *   (p)           (p)
				 *   / \           / \
				 *  N   S    -->  N   sl
				 *     / \           / \
				 *    sl  Sr        L   S
				 *    /\               / \
				 *   L  R             R  Sr

				 * Note: p might be red, and then both
				 * p and sl are red after rotation(which
				 * breaks property 4). This is fixed in
				 * Case 4 (in __rb_rotate_set_parents()
				 *         which set sl the color of p
				 *         and set p RB_BLACK)
				 *
				 *   (p)            (sl)
				 *   / \            /  \
				 *  N   sl   -->   P    S
				 *      /\        /\    /\
				 *     L  S      N  L  R  Sr
				 *       / \
				 *      R   Sr
				 */

				// tmp2: sl
				t.XReplaceN(tmp2, p)
				p.AddRight(tmp2.left)
				tmp2.AddLeft(p)

				s.AddLeft(tmp2.right)
				tmp2.AddRight(s)

				n = p
				p = tmp2
				tmp1 = s.right
			}
			/*
			 * Case 4 - left rotate at parent + color flips
			 * (p and sl could be either color here.
			 *  After rotation, p becomes black, s acquires
			 *  p's color, and sl keeps its color)
			 *
			 *      (p)             (s)
			 *      / \             / \
			 *     N   S     -->   P   Sr
			 *        / \         / \
			 *      (sl) sr      N  (sl)
			 */
			tmp2 = s.left
			t.XReplaceN(s, p)

			p.AddRight(tmp2)
			s.AddLeft(p)
			tmp1.isBlack = BLACK
			s.isBlack = p.isBlack
			p.isBlack = BLACK

		} else {
			// n = p.right
			s = p.left
			if !IsBlack(s) {
				/*
				 * Case 1 - left rotate at parent
				 *
				 *     P               S
				 *    / \             / \
				 *   s   N    -->    Sl   p
				 *  / \                  / \
				 * Sl  Sr               Sr  P
				 */
				tmp1 = s.right
				p.AddLeft(tmp1)
				t.XReplaceN(s, p)
				s.isBlack = BLACK
				s.AddRight(p)
				p.isBlack = RED
				s = tmp1
			}
			tmp1 = s.left
			if IsBlack(tmp1) {
				tmp2 = s.right
				if IsBlack(tmp2) {
					/*
					 * Case 2 - sibling color flip
					 * (p could be either color here)
					 *
					 *    (p)           (p)
					 *    / \           / \
					 *   S   N    -->  s   N
					 *  / \           / \
					 * Sl  Sr        Sl  Sr
					 *
					 * This leaves us violating 5) which
					 * can be fixed by flipping p to black
					 * if it was red, or by recursing at p.
					 * p is red when coming from Case 1.
					 */
					s.isBlack = RED
					if !p.isBlack {
						p.isBlack = BLACK
					} else {
						if p != t.root {
							n = p
							p = n.parent
							continue
						}
					}
					break
				}
				/*
				 * Case 3 - right rotate at sibling
				 * (p could be either color here)
				 *
				 *    (p)           (p)          (sr)
				 *    / \           / \          /  \
				 *   S   N    -->  sr   N  -->  S    P
				 *  / \           / \          / \  / \
				 * Sl  sr        S   R        Sl  L R  N
				 *     /\       / \
				 *    L  R     Sl  L

				 * Note: p might be red, and then both
				 * p and sl are red after rotation(which
				 * breaks property 4). This is fixed in
				 * Case 4 (in __rb_rotate_set_parents()
				 *         which set sl the color of p
				 *         and set p RB_BLACK)
				 */

				// tmp2: sl
				t.XReplaceN(tmp2, p)
				p.AddLeft(tmp2.right)
				tmp2.AddRight(p)

				s.AddRight(tmp2.left)
				tmp2.AddLeft(s)

				n = p
				p = tmp2
				tmp1 = s.left
			}
			/*
			 * Case 4 - left rotate at parent + color flips
			 * (p and sl could be either color here.
			 *  After rotation, p becomes black, s acquires
			 *  p's color, and sl keeps its color)
			 *
			 *      (p)             (s)
			 *      / \             / \
			 *     S   N     -->   Sl  P
			 *    / \                 / \
			 *   sl (sr)            (sr) N
			 */
			tmp2 = s.right
			t.XReplaceN(s, p)

			p.AddLeft(tmp2)
			s.AddRight(p)
			tmp1.isBlack = BLACK
			s.isBlack = p.isBlack
			p.isBlack = BLACK

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
	if x == t.root {
		t.root = c
		if c != nil {
			c.parent = nil
		}
		return true

	}

	if x == x.parent.left {
		x.parent.AddLeft(c)

	} else {
		x.parent.AddRight(c)
	}

	if x.isBlack {
		if IsBlack(c) {
			t.fixDelete(c, x.parent)
		} else {
			c.isBlack = BLACK
		}

	}

	return true
}

func IsBlack[K constraints.Ordered, V interface{}](x *Node[K, V]) Color {
	return x == nil || x.isBlack
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
