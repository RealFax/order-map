package odmap

import (
	"cmp"
	"encoding/json"
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
	json.Marshaler
}

type Map[K cmp.Ordered, V any] interface {
	internal[K, V]
	feature[K, V]
}
