package concurrency

import (
	"sync"

	"golang.org/x/exp/constraints"
)

type SaveMapNumeric[K constraints.Ordered, V Number] struct {
	m  map[K]V
	mx sync.Mutex
}

func NewSaveMapNumeric[K constraints.Ordered, V Number](ln int) *SaveMapNumeric[K, V] {
	return &SaveMapNumeric[K, V]{
		m: make(map[K]V, ln),
	}
}

func (m *SaveMapNumeric[K, V]) Set(k K, v V) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.m[k] = v
}

func (m *SaveMapNumeric[K, V]) Add(k K, v V) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.m[k] += v
}

func (m *SaveMapNumeric[K, V]) Ok(k K) bool {
	m.mx.Lock()
	defer m.mx.Unlock()
	_, ok := m.m[k]
	return ok
}

func (m *SaveMapNumeric[K, V]) Unsafe() map[K]V {
	m.mx.Lock()
	defer m.mx.Unlock()
	return m.m
}

func (m *SaveMapNumeric[K, V]) Len() int {
	m.mx.Lock()
	defer m.mx.Unlock()
	return len(m.m)
}
