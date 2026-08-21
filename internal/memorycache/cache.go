// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package memorycache

import (
	"sync"
)

type Storage[K comparable, V any] struct {
	mu    sync.Mutex
	items map[K]V
}

func New[K comparable, V any]() *Storage[K, V] {
	s := Storage[K, V]{
		items: make(map[K]V, 32),
	}

	return &s
}

func (s *Storage[K, V]) Get(key K) (*V, bool) {
	defer s.mu.Unlock()
	s.mu.Lock()

	v, ok := s.items[key]
	if !ok {
		return nil, false
	}

	return &v, true
}

func (s *Storage[K, V]) Set(key K, value V) {
	defer s.mu.Unlock()
	s.mu.Lock()

	s.items[key] = value
}

func (s *Storage[K, V]) Delete(key K) {
	defer s.mu.Unlock()
	s.mu.Lock()

	delete(s.items, key)
}
