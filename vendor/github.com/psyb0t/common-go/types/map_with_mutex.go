package commontypes

import (
	"sync"
)

type MapWithMutex[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewMapWithMutex[K comparable, V any]() *MapWithMutex[K, V] {
	return &MapWithMutex[K, V]{
		data: make(map[K]V),
	}
}

func (m *MapWithMutex[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
}

func (m *MapWithMutex[K, V]) Get(key K) (V, bool) { //nolint:ireturn
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.data[key]

	return value, exists
}

func (m *MapWithMutex[K, V]) Delete(key K) (V, bool) { //nolint:ireturn
	m.mu.Lock()
	defer m.mu.Unlock()

	value, exists := m.data[key]
	if exists {
		delete(m.data, key)
	}

	return value, exists
}

func (m *MapWithMutex[K, V]) Clear() map[K]V {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.data
	m.data = make(map[K]V)

	return old
}

func (m *MapWithMutex[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.data)
}
