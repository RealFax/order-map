// source: https://github.com/liyue201/gostl/blob/master/utils/iterator/iterator.go

package omap

import (
	"cmp"
	"sync/atomic"
)

type KvVisitor[K cmp.Ordered, V any] func(key K, value V) bool

// Color defines node color type
type Color bool

// Define node 's colors
const (
	RED   = false
	BLACK = true
)

// Node is a tree node
type Node[K cmp.Ordered, V any] struct {
	parent *Node[K, V]
	left   *Node[K, V]
	right  *Node[K, V]
	color  Color
	key    K
	value  *atomic.Pointer[V]
}

// Key returns node's key
func (n *Node[K, V]) Key() K {
	return n.key
}

// Value returns node's value
func (n *Node[K, V]) Value() V {
	p := n.value.Load()
	if p == nil {
		return empty[V]()
	}
	return *p
}

// SetValue sets node's value
func (n *Node[K, V]) SetValue(val V) {
	n.value.Store(&val)
}

// Next returns the Node's successor as an iterator.
func (n *Node[K, V]) Next() *Node[K, V] {
	return successor(n)
}

// Prev returns the Node's predecessor as an iterator.
func (n *Node[K, V]) Prev() *Node[K, V] {
	return presuccessor(n)
}

// successor returns the successor of the Node
func successor[K cmp.Ordered, V any](x *Node[K, V]) *Node[K, V] {
	if x.right != nil {
		return minimum(x.right)
	}
	y := x.parent
	for y != nil && x == y.right {
		x = y
		y = x.parent
	}
	return y
}

// presuccessor returns the presuccessor of the Node
func presuccessor[K cmp.Ordered, V any](x *Node[K, V]) *Node[K, V] {
	if x.left != nil {
		return maximum(x.left)
	}
	if x.parent != nil {
		if x.parent.right == x {
			return x.parent
		}
		for x.parent != nil && x.parent.left == x {
			x = x.parent
		}
		return x.parent
	}
	return nil
}

// minimum finds the minimum Node of subtree n.
func minimum[K cmp.Ordered, V any](n *Node[K, V]) *Node[K, V] {
	for n.left != nil {
		n = n.left
	}
	return n
}

// maximum finds the maximum Node of subtree n.
func maximum[K cmp.Ordered, V any](n *Node[K, V]) *Node[K, V] {
	for n.right != nil {
		n = n.right
	}
	return n
}

// getColor returns the node's color
func getColor[K cmp.Ordered, V any](n *Node[K, V]) Color {
	if n == nil {
		return BLACK
	}
	return n.color
}
