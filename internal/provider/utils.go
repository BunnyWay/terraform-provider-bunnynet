package provider

import (
	"golang.org/x/exp/constraints"
	"sync"
)

func mapValueToKey[T constraints.Integer](mapped map[T]string, value string) T {
	for k, v := range mapped {
		if v == value {
			return k
		}
	}

	panic("value not found in map")
}

func mapKeyToValue[T constraints.Integer](mapped map[T]string, value T) string {
	if v, ok := mapped[value]; ok {
		return v
	}

	panic("key not found in map")
}

// Mutex to manage concurrent changes to pullzone sub-resources (i.e. bunny_pullzone_optimizer_class)
// Based on https://discuss.hashicorp.com/t/cooping-with-parallelism-is-there-a-way-to-prioritise-resource-types/55690
var pzMutex *pullzoneMutex

func init() {
	pzMutex = &pullzoneMutex{mu: sync.Mutex{}, pullzones: map[int64]*sync.Mutex{}}
}

type pullzoneMutex struct {
	mu        sync.Mutex
	pullzones map[int64]*sync.Mutex
}

func (p *pullzoneMutex) Lock(id int64) {
	p.mu.Lock()
	if v, ok := p.pullzones[id]; ok {
		p.mu.Unlock()
		v.Lock()
	} else {
		p.pullzones[id] = &sync.Mutex{}
		p.mu.Unlock()
		p.pullzones[id].Lock()
	}
}

func (p *pullzoneMutex) Unlock(id int64) {
	p.mu.Lock()
	if _, ok := p.pullzones[id]; ok {
		p.pullzones[id].Unlock()
		delete(p.pullzones, id)
	}
	p.mu.Unlock()
}
