package rbtree

import (
	"testing"
)

type IntNode = Node[int, int]
type IntRBTree = RBTree[int, int]

func TestNode(t *testing.T) {
	black := IntNode{isBlack: true, key: 1, value: 0}
	t.Logf("Node: %#v\n", &black)
	red := IntNode{isBlack: false, key: 1, value: 0}
	t.Logf("Node: %#v\n", &red)
}

func TestRW(t *testing.T) {
	tree := IntRBTree{}
	m := make(map[int]int)
	m[0] = 1
	m[1] = 2
	m[2] = 3
	for k, v := range m {
		tree.Insert(k, v)
	}

	for k, v := range m {
		node := tree.GetNode(k)
		if node != nil {
			nk, nv := node.Pair()
			if nk != k {
				t.Fatalf("Wrong Key: %d -> %d\n", k, nk)
			}
			if nv != v {
				t.Fatalf("Wrong Val: Key=%d, %d -> %d", k, v, nv)
			}
		} else {
			t.Fatalf("Key Not Found %d\n", k)
		}
	}

}
