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
	Root *Node[K, V]
}

func (t *RBTree[K, V]) Insert(key K, value V) {
	var y *Node[K, V] = nil
	x := t.Root
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
		t.Root = &Node[K, V]{key: key, value: value, isBlack: BLACK}
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
	if g.left == p {
		u = g.right
	} else {
		u = g.left
	}

	if u != nil {
		if !u.isBlack {
			//      B       ->R
			//     /\        /\
			//    R  R  =>  B  B
			//   /         /
			//->R         R

			u.isBlack = BLACK
			p.isBlack = BLACK
			if g != t.Root {
				g.isBlack = RED
				t.InsertFix(g)
			}

			return
		}
	}

}

func (t *RBTree[K, V]) Delete(key K) bool {

	return false
}

func (t *RBTree[K, V]) Get(key K) (value V, ok bool) {
	node := t.Root
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
	node := t.Root
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
