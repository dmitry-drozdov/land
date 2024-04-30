package concurrency

import (
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type Map[K comparable, V Number] struct {
	parts  []*mapPart[K, V]
	shards int64
	hash   func(K) int64
}

func NewMap[K comparable, V Number](shards int64, hash func(K) int64) *Map[K, V] {
	mp := &Map[K, V]{
		parts:  make([]*mapPart[K, V], shards),
		shards: shards,
		hash:   hash,
	}
	for i := range mp.parts {
		mp.parts[i] = newMapPart[K, V]()
	}
	return mp
}

func (m *Map[K, V]) Add(k K, v V) {
	ptr := m.hash(k)
	if ptr < 0 || ptr >= m.shards {
		panic("incorrect hash")
	}
	mp := m.parts[ptr]
	mp.Add(k, v)
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	ptr := m.hash(k)
	if ptr < 0 || ptr >= m.shards {
		panic("incorrect hash")
	}
	mp := m.parts[ptr]
	return mp.Get(k)
}

func (m *Map[K, V]) Inc(k K) {
	ptr := m.hash(k)
	if ptr < 0 || ptr >= m.shards {
		panic("incorrect hash")
	}
	mp := m.parts[ptr]
	mp.Inc(k)
}

// for tests only!
func (m *Map[K, V]) Sum() V {
	var sum V
	for _, p := range m.parts {
		for _, v := range p.mp {
			sum += v
		}
	}
	return sum
}

func (m *Map[K, V]) Distribution() V {
	var sum V
	for _, p := range m.parts {
		fmt.Printf("%d+", len(p.mp))
	}
	return sum
}

type mapPart[K comparable, V Number] struct {
	mp map[K]V
	*sync.RWMutex
}

func newMapPart[K comparable, V Number]() *mapPart[K, V] {
	return &mapPart[K, V]{
		map[K]V{},
		&sync.RWMutex{},
	}
}

func (m *mapPart[K, V]) Add(k K, v V) {
	m.Lock()
	m.mp[k] = v
	m.Unlock()
}

func (m *mapPart[K, V]) Inc(k K) {
	m.Lock()
	m.mp[k]++
	m.Unlock()
}

func (m *mapPart[K, V]) Get(k K) (v V, ok bool) {
	m.RLock()
	v, ok = m.mp[k]
	m.mp[k]++
	m.RUnlock()
	return
}
