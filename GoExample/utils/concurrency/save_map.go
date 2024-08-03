package concurrency

import (
	"sync"

	"golang.org/x/exp/constraints"
)

type SaveMap[K constraints.Ordered, V any] struct {
	m  map[K]V
	mx sync.Mutex
}

func NewSaveMap[K constraints.Ordered, V any](ln int) *SaveMap[K, V] {
	return &SaveMap[K, V]{
		m: make(map[K]V, ln),
	}
}

func (m *SaveMap[K, V]) Set(k K, v V) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.m[k] = v
}

func (m *SaveMap[K, V]) Unsafe() map[K]V {
	m.mx.Lock()
	defer m.mx.Unlock()
	return m.m
}

func (m *SaveMap[K, V]) Len() int {
	m.mx.Lock()
	defer m.mx.Unlock()
	return len(m.m)
}
