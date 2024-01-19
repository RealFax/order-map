package odmap

import (
	"cmp"
	"fmt"
)

type internal[K cmp.Ordered, V any] interface {
	Load(K) (V, bool)
	Store(K, V)
	LoadOrStore(K, V) (V, bool)
	LoadAndDelete(K) (V, bool)
	Delete(K)
	Swap(K, V) (V, bool)
	CompareAndSwap(key K, old, new V) bool
	CompareAndDelete(K, V) bool
	Range(func(K, V) bool)
}

type feature[K cmp.Ordered, V any] interface {
	Len() int64
	Contains(K) bool
}

type Map[K cmp.Ordered, V any] interface {
	internal[K, V]
	feature[K, V]
}

type omap[K cmp.Ordered, V any] struct {
	tree *RBTree[K, V]
}

func (m *omap[K, V]) Load(key K) (V, bool) {
	node := m.tree.FindNode(key)
	if node == nil {
		return empty[V](), false
	}
	return node.Value(), true
}

func (m *omap[K, V]) Store(key K, value V) {
	_, _ = m.Swap(key, value)
}

func (m *omap[K, V]) Swap(key K, value V) (V, bool) {
	node := m.tree.FindNode(key)
	if node == nil {
		// node not found
		m.tree.Insert(key, value)
		return empty[V](), false
	}
	oldValue := node.Value()
	node.SetValue(value)
	return oldValue, true
}

func (m *omap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	node := m.tree.FindNode(key)
	if node != nil {
		return node.Value(), true
	}
	m.tree.Insert(key, value)
	return empty[V](), false
}

func (m *omap[K, V]) LoadAndDelete(key K) (V, bool) {
	node := m.tree.FindNode(key)
	if node != nil {
		m.tree.Delete(node)
		return node.Value(), true
	}
	return empty[V](), false
}

func (m *omap[K, V]) Delete(key K) {
	node := m.tree.FindNode(key)
	if node != nil {
		m.tree.Delete(node)
	}
}

func (m *omap[K, V]) CompareAndSwap(key K, old, new V) bool {
	node := m.tree.FindNode(key)
	if node == nil {
		return false
	}
	return node.value.CompareAndSwap(&old, &new)
}

func (m *omap[K, V]) CompareAndDelete(key K, old V) bool {
	node := m.tree.FindNode(key)
	if node == nil {
		return false
	}

	p := node.value.Load()
	if any(*p) != any(old) {
		return false
	}

	m.tree.Delete(node)
	return true
}

func (m *omap[K, V]) Range(fc func(key K, value V) bool) {
	for iter := m.tree.IterFirst(); iter.IsValid(); iter.Next() {
		fmt.Println(iter.Key())
		if !fc(iter.Key(), iter.Value()) {
			return
		}
	}
}

func (m *omap[K, V]) Len() int64 {
	return int64(m.tree.Size())
}

func (m *omap[K, V]) Contains(key K) bool {
	n := m.tree.FindNode(key)
	return n != nil
}

func newODMap[K cmp.Ordered, V any](opts ...Option[K, V]) *omap[K, V] {
	m := &omap[K, V]{tree: NewRBTree[K, V](func(a K, b K) int {
		return cmp.Compare(a, b)
	})}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func New[K cmp.Ordered, V any](opts ...Option[K, V]) Map[K, V] {
	return newODMap[K, V](opts...)
}
