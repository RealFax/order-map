// source: https://github.com/liyue201/gostl/blob/master/utils/iterator/iterator.go

package odmap

import (
	"cmp"
	"sync/atomic"
)

type KVisitor[K cmp.Ordered, V any] func(key K, value V) bool

// Color defines node color type
type Color bool

// Define node 's colors
const (
	RED   = false
	BLACK = true
)

// Entry is a tree entry
type Entry[K cmp.Ordered, V any] struct {
	expunged *V
	parent   *Entry[K, V]
	left     *Entry[K, V]
	right    *Entry[K, V]
	color    Color
	key      K
	value    *atomic.Pointer[V]
}

// Key returns node's key
func (n *Entry[K, V]) Key() K {
	return n.key
}

// Value returns node's value
func (n *Entry[K, V]) Value() V {
	p := n.value.Load()
	if p == nil {
		return empty[V]()
	}
	return *p
}

func (n *Entry[K, V]) load() (V, bool) {
	p := n.value.Load()
	if p == nil || p == n.expunged {
		return empty[V](), false
	}
	return *p, true
}

func (n *Entry[K, V]) tryCompareAndSwap(old, new V) bool {
	p := n.value.Load()
	if p == nil || p == n.expunged || any(*p) != any(old) {
		return false
	}

	nc := new
	for {
		if n.value.CompareAndSwap(p, &nc) {
			return true
		}

		p = n.value.Load()

		if p == nil || p == n.expunged || any(*p) != any(old) {
			return false
		}
	}
}

func (n *Entry[K, V]) unexpungeLocked() bool {
	return n.value.CompareAndSwap(n.expunged, nil)
}

func (n *Entry[K, V]) swapLocked(i *V) *V {
	return n.value.Swap(i)
}

func (n *Entry[K, V]) tryLoadOrStore(i V) (V, bool, bool) {
	p := n.value.Load()
	if p == n.expunged {
		return empty[V](), false, false
	}
	if p != nil {
		return *p, true, true
	}

	ic := i
	for {
		if n.value.CompareAndSwap(nil, &ic) {
			return i, false, true
		}

		p = n.value.Load()
		if p == n.expunged {
			return empty[V](), false, false
		}

		if p != nil {
			return *p, true, true
		}
	}
}

func (n *Entry[K, V]) delete() (V, bool) {
	for {
		p := n.value.Load()
		if p == nil || p == n.expunged {
			return empty[V](), false
		}

		if n.value.CompareAndSwap(p, nil) {
			return *p, true
		}
	}
}

func (n *Entry[K, V]) trySwap(i *V) (*V, bool) {
	for {
		p := n.value.Load()
		if p == n.expunged {
			return nil, false
		}
		if n.value.CompareAndSwap(p, i) {
			return p, true
		}
	}
}

func (n *Entry[K, V]) tryExpungeLocked() bool {
	p := n.value.Load()
	for p == nil {
		if n.value.CompareAndSwap(nil, n.expunged) {
			return true
		}
		p = n.value.Load()
	}

	return p == n.expunged
}

// ---- iterator ----

// Next returns the Entry's successor as an iterator.
func (n *Entry[K, V]) Next() *Entry[K, V] {
	return successor(n)
}

// Prev returns the Entry's predecessor as an iterator.
func (n *Entry[K, V]) Prev() *Entry[K, V] {
	return presuccessor(n)
}

// successor returns the successor of the Entry
func successor[K cmp.Ordered, V any](x *Entry[K, V]) *Entry[K, V] {
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

// presuccessor returns the presuccessor of the Entry
func presuccessor[K cmp.Ordered, V any](x *Entry[K, V]) *Entry[K, V] {
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

// minimum finds the minimum Entry of subtree n.
func minimum[K cmp.Ordered, V any](n *Entry[K, V]) *Entry[K, V] {
	for n.left != nil {
		n = n.left
	}
	return n
}

// maximum finds the maximum Entry of subtree n.
func maximum[K cmp.Ordered, V any](n *Entry[K, V]) *Entry[K, V] {
	for n.right != nil {
		n = n.right
	}
	return n
}

// getColor returns the node's color
func getColor[K cmp.Ordered, V any](n *Entry[K, V]) Color {
	if n == nil {
		return BLACK
	}
	return n.color
}

func empty[V any]() V {
	var e V
	return e
}

func newPointerValue[V any](val V) *atomic.Pointer[V] {
	p := &atomic.Pointer[V]{}
	p.Store(&val)
	return p
}
