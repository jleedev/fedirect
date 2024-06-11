package main

import "sync"

type RWMap[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

func NewRWMap[K comparable, V any]() RWMap[K, V] {
	return RWMap[K, V]{
		make(map[K]V),
		sync.RWMutex{},
	}
}

func (r *RWMap[K, V]) Get(k K) (V, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result, ok := r.data[k]
	return result, ok
}

func (r *RWMap[K, V]) Set(k K, v V) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[k] = v
}
