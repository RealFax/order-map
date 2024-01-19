// source: https://github.com/liyue201/gostl/blob/master/utils/iterator/iterator.go

package odmap

import (
	"cmp"
)

// RBTree is a kind of self-balancing binary search tree in computer science.
// Each node of the binary tree has an extra bit, and that bit is often interpreted
// as the color (red or black) of the node. These color bits are used to ensure the tree
// remains approximately balanced during insertions and deletions.
type RBTree[K cmp.Ordered, V any] struct {
	size    int
	root    *Node[K, V]
	compare func(K, K) int
}

// Clear clears the RBTree
func (t *RBTree[K, V]) Clear() {
	t.root = nil
	t.size = 0
}

// FindNode the first node that the key is equal to the passed key and return it
func (t *RBTree[K, V]) FindNode(key K) *Node[K, V] {
	return t.findFirstNode(key)
}

// Begin returns the node with minimum key in the RBTree
func (t *RBTree[K, V]) Begin() *Node[K, V] {
	return t.First()
}

// First returns the node with minimum key in the RBTree
func (t *RBTree[K, V]) First() *Node[K, V] {
	if t.root == nil {
		return nil
	}
	return minimum(t.root)
}

// RBegin returns the Node with maximum key in the RBTree
func (t *RBTree[K, V]) RBegin() *Node[K, V] {
	return t.Last()
}

// Last returns the Node with maximum key in the RBTree
func (t *RBTree[K, V]) Last() *Node[K, V] {
	if t.root == nil {
		return nil
	}
	return maximum(t.root)
}

// IterFirst returns the iterator of first node
func (t *RBTree[K, V]) IterFirst() *RBTreeIterator[K, V] {
	return NewIterator(t.First())
}

// IterLast returns the iterator of first node
func (t *RBTree[K, V]) IterLast() *RBTreeIterator[K, V] {
	return NewIterator(t.Last())
}

// Empty returns true if Tree is empty,otherwise returns false.
func (t *RBTree[K, V]) Empty() bool {
	return t.size == 0
}

// Size returns the size of the rbtree.
func (t *RBTree[K, V]) Size() int {
	return t.size
}

// Insert inserts a key-value pair into the RBTree.
func (t *RBTree[K, V]) Insert(key K, value V) {
	x := t.root
	var y *Node[K, V]

	for x != nil {
		y = x
		if t.compare(key, x.key) < 0 {
			x = x.left
		} else {
			x = x.right
		}
	}

	z := &Node[K, V]{parent: y, color: RED, key: key, value: newPointerValue(value)}
	t.size++

	if y == nil {
		z.color = BLACK
		t.root = z
		return
	} else if t.compare(z.key, y.key) < 0 {
		y.left = z
	} else {
		y.right = z
	}
	t.rbInsertFixup(z)
}

func (t *RBTree[K, V]) rbInsertFixup(z *Node[K, V]) {
	var y *Node[K, V]
	for z.parent != nil && !z.parent.color {
		if z.parent == z.parent.parent.left {
			y = z.parent.parent.right
			if y != nil && !y.color {
				z.parent.color = BLACK
				y.color = BLACK
				z.parent.parent.color = RED
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					t.leftRotate(z)
				}
				z.parent.color = BLACK
				z.parent.parent.color = RED
				t.rightRotate(z.parent.parent)
			}
		} else {
			y = z.parent.parent.left
			if y != nil && !y.color {
				z.parent.color = BLACK
				y.color = BLACK
				z.parent.parent.color = RED
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					t.rightRotate(z)
				}
				z.parent.color = BLACK
				z.parent.parent.color = RED
				t.leftRotate(z.parent.parent)
			}
		}
	}
	t.root.color = BLACK
}

// Delete deletes node from the RBTree
func (t *RBTree[K, V]) Delete(node *Node[K, V]) {
	z := node
	if z == nil {
		return
	}

	var x, y *Node[K, V]
	if z.left != nil && z.right != nil {
		y = successor(z)
	} else {
		y = z
	}

	if y.left != nil {
		x = y.left
	} else {
		x = y.right
	}

	xparent := y.parent
	if x != nil {
		x.parent = xparent
	}
	if y.parent == nil {
		t.root = x
	} else if y == y.parent.left {
		y.parent.left = x
	} else {
		y.parent.right = x
	}

	if y != z {
		z.key = y.key
		z.value.Store(y.value.Load())
	}

	if y.color {
		t.rbDeleteFixup(x, xparent)
	}
	t.size--
}

func (t *RBTree[K, V]) rbDeleteFixup(x, parent *Node[K, V]) {
	var w *Node[K, V]
	for x != t.root && getColor(x) {
		if x != nil {
			parent = x.parent
		}
		if x == parent.left {
			x, w = t.rbFixupLeft(x, parent, w)
		} else {
			x, w = t.rbFixupRight(x, parent, w)
		}
	}
	if x != nil {
		x.color = BLACK
	}
}

func (t *RBTree[K, V]) rbFixupLeft(x, parent, w *Node[K, V]) (*Node[K, V], *Node[K, V]) {
	w = parent.right
	if !w.color {
		w.color = BLACK
		parent.color = RED
		t.leftRotate(parent)
		w = parent.right
	}
	if getColor(w.left) && getColor(w.right) {
		w.color = RED
		x = parent
	} else {
		if getColor(w.right) {
			if w.left != nil {
				w.left.color = BLACK
			}
			w.color = RED
			t.rightRotate(w)
			w = parent.right
		}
		w.color = parent.color
		parent.color = BLACK
		if w.right != nil {
			w.right.color = BLACK
		}
		t.leftRotate(parent)
		x = t.root
	}
	return x, w
}

func (t *RBTree[K, V]) rbFixupRight(x, parent, w *Node[K, V]) (*Node[K, V], *Node[K, V]) {
	w = parent.left
	if !w.color {
		w.color = BLACK
		parent.color = RED
		t.rightRotate(parent)
		w = parent.left
	}
	if getColor(w.left) && getColor(w.right) {
		w.color = RED
		x = parent
	} else {
		if getColor(w.left) {
			if w.right != nil {
				w.right.color = BLACK
			}
			w.color = RED
			t.leftRotate(w)
			w = parent.left
		}
		w.color = parent.color
		parent.color = BLACK
		if w.left != nil {
			w.left.color = BLACK
		}
		t.rightRotate(parent)
		x = t.root
	}
	return x, w
}

func (t *RBTree[K, V]) leftRotate(x *Node[K, V]) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		t.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

func (t *RBTree[K, V]) rightRotate(x *Node[K, V]) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		t.root = y
	} else if x == x.parent.right {
		x.parent.right = y
	} else {
		x.parent.left = y
	}
	y.right = x
	x.parent = y
}

// findNode finds the node that its key is equal to the passed key, and returns it.
func (t *RBTree[K, V]) findNode(key K) *Node[K, V] {
	x := t.root
	for x != nil {
		if t.compare(key, x.key) < 0 {
			x = x.left
		} else {
			if t.compare(key, x.key) == 0 {
				return x
			}
			x = x.right
		}
	}
	return nil
}

// findNode finds the first node that its key is equal to the passed key, and returns it
func (t *RBTree[K, V]) findFirstNode(key K) *Node[K, V] {
	node := t.FindLowerBoundNode(key)
	if node == nil {
		return nil
	}
	if t.compare(node.key, key) == 0 {
		return node
	}
	return nil
}

// FindLowerBoundNode finds the first node that its key is equal or greater than the passed key, and returns it
func (t *RBTree[K, V]) FindLowerBoundNode(key K) *Node[K, V] {
	return t.findLowerBoundNode(t.root, key)
}

func (t *RBTree[K, V]) findLowerBoundNode(x *Node[K, V], key K) *Node[K, V] {
	if x == nil {
		return nil
	}
	if t.compare(key, x.key) <= 0 {
		ret := t.findLowerBoundNode(x.left, key)
		if ret == nil {
			return x
		}
		if t.compare(ret.key, x.key) <= 0 {
			return ret
		}
		return x
	}
	return t.findLowerBoundNode(x.right, key)
}

// FindUpperBoundNode finds the first node that its key is greater than the passed key, and returns it
func (t *RBTree[K, V]) FindUpperBoundNode(key K) *Node[K, V] {
	return t.findUpperBoundNode(t.root, key)
}

func (t *RBTree[K, V]) findUpperBoundNode(x *Node[K, V], key K) *Node[K, V] {
	if x == nil {
		return nil
	}
	if t.compare(key, x.key) >= 0 {
		return t.findUpperBoundNode(x.right, key)
	}
	ret := t.findUpperBoundNode(x.left, key)
	if ret == nil {
		return x
	}
	if t.compare(ret.key, x.key) <= 0 {
		return ret
	}
	return x
}

// Traversal traversals elements in the RBTree, it will not stop until to the end of RBTree or the visitor returns false
func (t *RBTree[K, V]) Traversal(visitor KVisitor[K, V]) {
	for node := t.First(); node != nil; node = node.Next() {
		if !visitor(node.key, node.Value()) {
			break
		}
	}
}

// NewRBTree creates a new RBTree
func NewRBTree[K cmp.Ordered, V any](comparer func(K, K) int) *RBTree[K, V] {
	return &RBTree[K, V]{
		compare: comparer,
	}
}
