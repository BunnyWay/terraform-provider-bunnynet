package utils

import (
	"sync"
)

func NewIdMutex() IdMutex {
	return IdMutex{
		mu:  sync.Mutex{},
		ids: make(map[int64]*sync.Mutex, 8),
	}
}

type IdMutex struct {
	mu  sync.Mutex
	ids map[int64]*sync.Mutex
}

func (p *IdMutex) Lock(id int64) {
	p.mu.Lock()
	if v, ok := p.ids[id]; ok {
		p.mu.Unlock()
		v.Lock()
	} else {
		p.ids[id] = &sync.Mutex{}
		p.ids[id].Lock()
		p.mu.Unlock()
	}
}

func (p *IdMutex) Unlock(id int64) {
	p.mu.Lock()
	if _, ok := p.ids[id]; ok {
		p.ids[id].Unlock()
	}
	p.mu.Unlock()
}
