//go:build safety_map

package odmap

import (
	"cmp"
	"encoding/json"
	"sync"
	"sync/atomic"
)

type readonly[K cmp.Ordered, V any] struct {
	m       *RBTree[K, V]
	amended bool
}

type safetyMap[K cmp.Ordered, V any] struct {
	compare func(K, K) int

	mu sync.Mutex

	read atomic.Pointer[readonly[K, V]]

	dirty *RBTree[K, V]

	misses int
}

func (m *safetyMap[K, V]) loadReadonly() readonly[K, V] {
	if p := m.read.Load(); p != nil {
		return *p
	}
	return readonly[K, V]{}
}

func (m *safetyMap[K, V]) missLocked() {
	m.misses++
	if m.misses < m.dirty.Size() {
		return
	}
	m.read.Store(&readonly[K, V]{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

func (m *safetyMap[K, V]) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read := m.loadReadonly()
	m.dirty = NewRBTree[K, V](m.compare)

	for iter := read.m.IterFirst(); iter.IsValid(); iter.Next() {
		if !iter.node.tryExpungeLocked() {
			m.dirty.put(iter.node.Key(), iter.node.Value())
		}
	}
}

func (m *safetyMap[K, V]) Load(key K) (V, bool) {
	read := m.loadReadonly()
	e, ok := read.m.get(key)
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadonly()
		e, ok = read.m.get(key)
		if !ok && read.amended {
			e, ok = m.dirty.get(key)
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return empty[V](), false
	}
	return e.load()
}

func (m *safetyMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	read := m.loadReadonly()
	if e, ok := read.m.get(key); ok {
		if v, ok := e.trySwap(&value); ok {
			if v == nil {
				return empty[V](), false
			}
			return *v, true
		}
	}

	m.mu.Lock()
	read = m.loadReadonly()
	if e, ok := read.m.get(key); ok {
		if e.unexpungeLocked() {
			m.dirty.put(key, e.Value())
		}

		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else if e, ok := m.dirty.get(key); ok {
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else {
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(&readonly[K, V]{m: read.m, amended: true})
		}
		m.dirty.put(key, value)
	}
	m.mu.Unlock()
	return previous, loaded
}

func (m *safetyMap[K, V]) Store(key K, value V) {
	_, _ = m.Swap(key, value)
}

func (m *safetyMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	read := m.loadReadonly()
	if e, ok := read.m.get(key); ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	m.mu.Lock()
	read = m.loadReadonly()
	if e, ok := read.m.get(key); ok {
		if e.unexpungeLocked() {
			m.dirty.put(key, e.Value())
		}
		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := m.dirty.get(key); ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		m.missLocked()
	} else {
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(&readonly[K, V]{m: read.m, amended: true})
		}
		m.dirty.put(key, value)
		actual, loaded = value, false
	}
	m.mu.Unlock()

	return actual, loaded
}

func (m *safetyMap[K, V]) LoadAndDelete(key K) (V, bool) {
	read := m.loadReadonly()
	e, ok := read.m.get(key)
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadonly()
		e, ok = read.m.get(key)
		if !ok && read.amended {
			e, ok = m.dirty.get(key)
			m.dirty.del(key)
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if ok {
		return e.delete()
	}
	return empty[V](), false
}

func (m *safetyMap[K, V]) Delete(key K) {
	_, _ = m.LoadAndDelete(key)
}

func (m *safetyMap[K, V]) CompareAndSwap(key K, old, new V) (swapped bool) {
	read := m.loadReadonly()
	if e, ok := read.m.get(key); ok {
		return e.tryCompareAndSwap(old, new)
	} else if !read.amended {
		return false
	}

	m.mu.Lock()
	read = m.loadReadonly()
	if e, ok := read.m.get(key); ok {
		swapped = e.tryCompareAndSwap(old, new)
	} else if e, ok := m.dirty.get(key); ok {
		swapped = e.tryCompareAndSwap(old, new)
		m.missLocked()
	}
	m.mu.Unlock()

	return swapped
}

func (m *safetyMap[K, V]) CompareAndDelete(key K, old V) bool {
	read := m.loadReadonly()
	e, ok := read.m.get(key)
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadonly()
		e, ok = read.m.get(key)
		if !ok && read.amended {
			e, ok = m.dirty.get(key)
			m.missLocked()
		}
		m.mu.Unlock()
	}

	for ok {
		p := e.value.Load()
		if p == nil || p == e.expunged || any(*p) != any(old) {
			return false
		}

		if e.value.CompareAndSwap(p, nil) {
			return true
		}
	}
	return false
}

func (m *safetyMap[K, V]) Range(fc func(key K, value V) bool) {
	read := m.loadReadonly()
	if read.amended {
		m.mu.Lock()
		read = m.loadReadonly()
		if read.amended {
			read = readonly[K, V]{m: m.dirty}
			m.read.Store(&read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for iter := read.m.IterFirst(); iter.IsValid(); iter.Next() {
		v, ok := iter.node.load()
		if !ok {
			continue
		}

		if !fc(iter.Key(), v) {
			break
		}
	}
}

func (m *safetyMap[K, V]) Len() int64 { return 0 }
func (m *safetyMap[K, V]) Contains(key K) bool {
	_, found := m.Load(key)
	return found
}

func (m *safetyMap[K, V]) MarshalJSON() ([]byte, error) {
	s := make([]Pair[K, V], 0, 1024)
	m.Range(func(key K, value V) bool {
		s = append(s, Pair[K, V]{Key: key, Value: value})
		return true
	})
	return json.Marshal(s)
}

func New[K cmp.Ordered, V any](opts ...Option[K, V]) Map[K, V] {
	m := &safetyMap[K, V]{}

	for _, opt := range opts {
		opt(m)
	}

	if m.compare == nil {
		m.compare = func(a K, b K) int {
			return cmp.Compare(a, b)
		}
	}

	m.read.Store(&readonly[K, V]{m: NewRBTree[K, V](m.compare), amended: true})
	m.dirty = NewRBTree[K, V](m.compare)

	return m
}
