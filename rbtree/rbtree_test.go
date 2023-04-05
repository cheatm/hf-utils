package rbtree

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

func Make1234(t *testing.T, tree *IntRBTree) {

	// Uncle, Parent: [RED]
	// GrandParent: (BLACK)
	//
	//       (3)       ->[3]    Root(3)
	//       / \         / \        / \
	//     [2] [4]  => (2) (4) => (2) (4)
	//     /           /          /
	// ->[1]         [1]        [1]

	tree.Insert(3, 0)
	tree.Insert(2, 0)
	tree.Insert(4, 0)

	assert.Equal(t, 3, tree.root.key, "Root key not match")
	assert.Equal(t, BLACK, tree.root.isBlack, "Root is not BLACK")
	assert.Equal(t, 2, tree.root.left.key, "Root.left key not match")
	assert.Equal(t, RED, tree.root.left.isBlack, "Root.left is not RED")
	assert.Equal(t, 4, tree.root.right.key, "Root.right key not match")
	assert.Equal(t, RED, tree.root.right.isBlack, "Root.right is not RED")

	tree.Insert(1, 0)

	assert.Equal(t, 3, tree.root.key, "Root key not match")
	assert.Equal(t, BLACK, tree.root.isBlack, "Root is not BLACK")
	assert.Equal(t, 2, tree.root.left.key, "Root.left key not match")
	assert.Equal(t, BLACK, tree.root.left.isBlack, "Root.left is not BLACK")
	assert.Equal(t, 4, tree.root.right.key, "Root.right key not match")
	assert.Equal(t, BLACK, tree.root.right.isBlack, "Root.right is not BLACK")
	assert.Equal(t, 1, tree.root.left.left.key, "Root.right key not match")
	assert.Equal(t, RED, tree.root.left.left.isBlack, "Root.right is not BLACK")

}

func TestFixRRB(t *testing.T) {
	tree := &IntRBTree{}

	Make1234(t, tree)
}

func assertNode(t *testing.T, node *IntNode, key int, color Color) {
	assert.Equal(t, key, node.key, "Key")
	assert.Equal(t, color, node.isBlack, "Color")
}

func TestFixLL(t *testing.T) {
	// LL:

	// Root(6)
	//     / \
	//   [5] [7]

	tree := &IntRBTree{}
	tree.Insert(7, 7)
	tree.Insert(6, 6)
	tree.Insert(5, 5)

	CheckTree(
		t, tree,
		[]int{5, 6, 7},
		[]Color{RED, BLACK, RED},
	)

	// Root(6)
	//     / \
	//   (5) (7)
	//   /
	// [4]

	tree.Insert(4, 4)

	CheckTree(
		t, tree,
		[]int{4, 5, 6, 7},
		[]Color{RED, BLACK, BLACK, BLACK},
	)

	// n4 := n5.left
	// assertNode(t, n4, 4, RED)
	// assertNode(t, n5, 5, BLACK)
	// assertNode(t, n7, 7, BLACK)
	// assertNode(t, n6, 6, BLACK)

	//   Root(6)
	//       / \
	//     (4) (7)
	//     / \
	//   [3] [5]

	tree.Insert(3, 3)
	CheckTree(
		t, tree,
		[]int{3, 4, 5, 6, 7},
		[]Color{RED, BLACK, RED, BLACK, BLACK},
	)
	// n3 := n4.left
	// assertNode(t, n4, 4, BLACK)
	// assertNode(t, n3, 3, RED)
	// assertNode(t, n5, 5, RED)

	//   Root(6)
	//       / \
	//     [4] (7)
	//     / \
	//   (3) (5)
	//   /
	// [2]

	tree.Insert(2, 2)
	CheckTree(
		t, tree,
		[]int{2, 3, 4, 5, 6, 7},
		[]Color{RED, BLACK, RED, BLACK, BLACK, BLACK},
	)
	// n2 := n3.left

	// assertNode(t, n4, 4, RED)
	// assertNode(t, n3, 3, BLACK)
	// assertNode(t, n5, 5, BLACK)
	// assertNode(t, n2, 2, RED)

	//   Root(6)
	//       / \
	//     [4] (7)
	//     / \
	//   (2) (5)
	//   / \
	// [1] [3]
	tree.Insert(1, 1)
	CheckTree(
		t, tree,
		[]int{1, 2, 3, 4, 5, 6, 7},
		[]Color{RED, BLACK, RED, RED, BLACK, BLACK, BLACK},
	)

	// n1 := n2.left
	// assertNode(t, n4, 4, RED)
	// assertNode(t, n3, 3, RED)
	// assertNode(t, n1, 1, RED)
	// assertNode(t, n2, 2, BLACK)

	//   (1) (3)
	//   /
	// [0]

	tree.Insert(0, 0)

	assertNode(t, tree.root, 4, BLACK)
	// CheckTree(
	// 	t, tree,
	// 	[]int{0, 1, 2, 3, 4, 5, 6, 7},
	// 	[]Color{RED, BLACK, RED, BLACK, BLACK, BLACK, RED, BLACK},
	// )

}

func CheckTree(t *testing.T, tree *IntRBTree, keys []int, colors []Color) {
	root := tree.root
	if root == nil {
		return
	}

	assert.Equal(t, BLACK, root.isBlack, "Root should be black")
	index := 0
	MidCheckNode(t, &index, root, keys, colors)
}

func MidCheckNode(t *testing.T, index *int, node *IntNode, keys []int, colors []Color) {
	if node.left != nil {
		MidCheckNode(t, index, node.left, keys, colors)
	}

	assert.Equal(t, keys[*index], node.key, fmt.Sprintf("Key of node %d", *index))
	assert.Equal(t, colors[*index], node.isBlack, fmt.Sprintf("Color of node %d", *index))

	*index = *index + 1

	if node.right != nil {
		MidCheckNode(t, index, node.right, keys, colors)
	}

}
