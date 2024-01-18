// source: https://github.com/liyue201/gostl/blob/master/utils/iterator/iterator.go

package omap

import "cmp"

// ConstIterator is an interface of const iterator
type ConstIterator[T any] interface {
	IsValid() bool
	Next() ConstIterator[T]
	Value() T
	Clone() ConstIterator[T]
	Equal(other ConstIterator[T]) bool
}

// ConstBidIterator is an interface of const bidirectional iterator
type ConstBidIterator[T any] interface {
	ConstIterator[T]
	Prev() ConstBidIterator[T]
}

// RBTreeIterator is an iterator implementation of RBTree
type RBTreeIterator[K cmp.Ordered, V any] struct {
	node *Node[K, V]
}

// NewIterator creates a RBTreeIterator from the passed node
func NewIterator[K cmp.Ordered, V any](node *Node[K, V]) *RBTreeIterator[K, V] {
	return &RBTreeIterator[K, V]{node: node}
}

// IsValid returns true if the iterator is valid, otherwise returns false
func (iter *RBTreeIterator[K, V]) IsValid() bool {
	return iter.node != nil
}

// Next moves the pointer of the iterator to the next node, and returns itself
func (iter *RBTreeIterator[K, V]) Next() ConstIterator[V] {
	if iter.IsValid() {
		iter.node = iter.node.Next()
	}
	return iter
}

// Prev moves the pointer of the iterator to the previous node, and returns itself
func (iter *RBTreeIterator[K, V]) Prev() ConstBidIterator[V] {
	if iter.IsValid() {
		iter.node = iter.node.Prev()
	}
	return iter
}

// Key returns the node's key of the iterator point to
func (iter *RBTreeIterator[K, V]) Key() K {
	return iter.node.Key()
}

// Value returns the node's value of the iterator point to
func (iter *RBTreeIterator[K, V]) Value() V {
	return iter.node.Value()
}

// SetValue sets the node's value of the iterator point to
func (iter *RBTreeIterator[K, V]) SetValue(val V) error {
	iter.node.SetValue(val)
	return nil
}

// Clone clones the iterator into a new RBTreeIterator
func (iter *RBTreeIterator[K, V]) Clone() ConstIterator[V] {
	return NewIterator(iter.node)
}

// Equal returns true if the iterator is equal to the passed iterator
func (iter *RBTreeIterator[K, V]) Equal(other ConstIterator[V]) bool {
	otherIter, ok := other.(*RBTreeIterator[K, V])
	if !ok {
		return false
	}
	if otherIter.node == iter.node {
		return true
	}
	return false
}
