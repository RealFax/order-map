package omap

import "cmp"

type Option[K cmp.Ordered, V any] func(m *omap[K, V])

func WithComparer[K cmp.Ordered, V any](comparer func(K, K) int) Option[K, V] {
	return func(m *omap[K, V]) {
		m.tree.compare = comparer
	}
}
