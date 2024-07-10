package provider

import (
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
	"strconv"
	"strings"
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

func sliceMap[T any, R any](s []T, f func(v T) R) []R {
	var result []R
	for _, v := range s {
		result = append(result, f(v))
	}
	return result
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

func convertTimestampToSeconds(timestamp string) (uint64, error) {
	parts := strings.Split(timestamp, ":")
	if len(parts) != 2 {
		return 0, errors.New("Invalid timestamp format, expected \"00:00\"")
	}

	minutes, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return uint64(seconds + minutes*60), nil
}

func convertSecondsToTimestamp(seconds uint64) string {
	minutes := math.Floor(float64(seconds / 60))
	remainder := math.Mod(float64(seconds), 60)

	return fmt.Sprintf("%02d:%02d", int64(minutes), int64(remainder))
}
