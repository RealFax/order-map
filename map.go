// implementation of high performance concurrent safe ordered map

package seqMap

import (
	"sort"
	"sync"
	"sync/atomic"
)

type Value[V any] struct {
	Seq   int32
	Value V
}

type KV[K comparable, V any] struct {
	Key   K
	Value *atomic.Pointer[Value[V]]
}

type KVs[K comparable, V any] []KV[K, V]

func (s KVs[K, V]) Len() int           { return len(s) }
func (s KVs[K, V]) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s KVs[K, V]) Less(i, j int) bool { return s[i].Value.Load().Seq < s[j].Value.Load().Seq }

type SeqMap[K comparable, V any] struct {
	offset int32
	mu     sync.Mutex
	dirty  map[K]*atomic.Pointer[Value[V]]
}

func (s *SeqMap[K, V]) Store(key K, value V) {
	s.mu.Lock()
	ptr, ok := s.dirty[key]
	if ok {
		oldVal := ptr.Load()
		ptr.Swap(&Value[V]{Seq: oldVal.Seq, Value: value})

		s.mu.Unlock()
		return
	}

	atomic.AddInt32(&s.offset, 1)
	val := atomic.Pointer[Value[V]]{}
	val.Store(&Value[V]{Seq: s.offset, Value: value})
	s.dirty[key] = &val

	s.mu.Unlock()
}

func (s *SeqMap[K, V]) Load(key K) (V, bool) {
	s.mu.Lock()
	ptr, ok := s.dirty[key]
	s.mu.Unlock()
	if !ok {
		var zero V
		return zero, false
	}
	val := ptr.Load()
	return val.Value, true
}

func (s *SeqMap[K, V]) Delete(key K) {
	s.mu.Lock()
	ptr, ok := s.dirty[key]
	if !ok {
		s.mu.Unlock()
		return
	}
	atomic.AddInt32(&s.offset, -1)
	ptr.Store(nil)
	ptr = nil

	delete(s.dirty, key)
	s.mu.Unlock()
}

func (s *SeqMap[K, V]) Range(f func(key K, value V) bool) {
	s.mu.Lock()

	vs := make(KVs[K, V], atomic.LoadInt32(&s.offset))
	vsi := 0
	for key, val := range s.dirty {
		vs[vsi] = KV[K, V]{
			Key:   key,
			Value: val,
		}
		vsi++
	}

	sort.Sort(vs)

	for i := 0; i < int(atomic.LoadInt32(&s.offset)); i++ {
		ptr := vs[i].Value.Load()
		if !f(vs[i].Key, ptr.Value) {
			break
		}
	}

	s.mu.Unlock()
}

func (s *SeqMap[K, V]) DisorderedRange(f func(key K, value V) bool) {
	s.mu.Lock()
	for key, val := range s.dirty {
		if !f(key, val.Load().Value) {
			break
		}
	}
	s.mu.Unlock()
}

func (s *SeqMap[K, V]) Map() map[K]V {
	m := make(map[K]V)
	s.DisorderedRange(func(key K, value V) bool {
		m[key] = value
		return true
	})
	return m
}

func NewSeqMap[K comparable, V any]() *SeqMap[K, V] {
	return &SeqMap[K, V]{dirty: make(map[K]*atomic.Pointer[Value[V]])}
}
