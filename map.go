// implementation of high performance concurrent safe ordered map

package orderMap

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

type Map[K comparable, V any] struct {
	size  int32
	mu    sync.RWMutex
	dirty map[K]*atomic.Pointer[Value[V]]
}

func (s *Map[K, V]) Size() int32 {
	return atomic.LoadInt32(&s.size)
}

func (s *Map[K, V]) Store(key K, value V) {
	s.mu.Lock()
	ptr, ok := s.dirty[key]
	if ok {
		oldVal := ptr.Load()
		ptr.Swap(&Value[V]{Seq: oldVal.Seq, Value: value})

		s.mu.Unlock()
		return
	}

	atomic.AddInt32(&s.size, 1)
	val := atomic.Pointer[Value[V]]{}
	val.Store(&Value[V]{Seq: s.size, Value: value})
	s.dirty[key] = &val

	s.mu.Unlock()
}

func (s *Map[K, V]) Load(key K) (V, bool) {
	s.mu.RLock()
	ptr, ok := s.dirty[key]
	s.mu.RUnlock()
	if !ok {
		var zero V
		return zero, false
	}
	val := ptr.Load()
	return val.Value, true
}

func (s *Map[K, V]) Delete(key K) {
	s.mu.Lock()
	ptr, ok := s.dirty[key]
	if !ok {
		s.mu.Unlock()
		return
	}
	atomic.AddInt32(&s.size, -1)
	ptr.Store(nil)
	ptr = nil

	delete(s.dirty, key)
	s.mu.Unlock()
}

func (s *Map[K, V]) Range(f func(key K, value V) bool) {
	s.mu.RLock()

	vs := make(KVs[K, V], atomic.LoadInt32(&s.size))
	vsi := 0
	for key, val := range s.dirty {
		vs[vsi] = KV[K, V]{
			Key:   key,
			Value: val,
		}
		vsi++
	}

	sort.Sort(vs)

	for i := 0; i < int(atomic.LoadInt32(&s.size)); i++ {
		ptr := vs[i].Value.Load()
		if !f(vs[i].Key, ptr.Value) {
			break
		}
	}

	s.mu.RUnlock()
}

func (s *Map[K, V]) DisorderedRange(f func(key K, value V) bool) {
	s.mu.RLock()
	for key, val := range s.dirty {
		if !f(key, val.Load().Value) {
			break
		}
	}
	s.mu.RUnlock()
}

func (s *Map[K, V]) Map() map[K]V {
	m := make(map[K]V)
	s.DisorderedRange(func(key K, value V) bool {
		m[key] = value
		return true
	})
	return m
}

func (s *Map[K, V]) Reset() {
	s.mu.Lock()
	atomic.StoreInt32(&s.size, 0)
	s.dirty = nil
	s.dirty = make(map[K]*atomic.Pointer[Value[V]])
	s.mu.Unlock()
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{dirty: make(map[K]*atomic.Pointer[Value[V]])}
}
