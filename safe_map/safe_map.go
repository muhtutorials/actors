package safe_map

import (
	"fmt"
	"sync"
)

type SafeMap[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

// New creates a new thread safe map with optional size parameter.
func New[K comparable, V any](size ...int) *SafeMap[K, V] {
	var data map[K]V
	if len(size) > 0 {
		data = make(map[K]V, size[0])
	} else {
		data = make(map[K]V)
	}
	return &SafeMap[K, V]{
		data: data,
	}
}

func (m *SafeMap[K, V]) Insert(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SafeMap[K, V]) Get(key K) (V, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	if !ok {
		return value, fmt.Errorf("key '%v' not found", key)
	}
	return value, nil
}

func (m *SafeMap[K, V]) Update(key K, value V) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	if !ok {
		return fmt.Errorf("key '%v' not found", key)
	}
	m.data[key] = value
	return nil
}

func (m *SafeMap[K, V]) Delete(key K) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	if !ok {
		return fmt.Errorf("key '%v' not found", key)
	}
	delete(m.data, key)
	return nil
}

func (m *SafeMap[K, V]) HasKey(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *SafeMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func (m *SafeMap[K, V]) ForEach(fn func(K, V)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.data {
		fn(k, v)
	}
}
