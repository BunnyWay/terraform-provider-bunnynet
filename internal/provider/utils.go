// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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

// Mutex to manage concurrent changes to pullzone sub-resources (i.e. bunnynet_pullzone_edgerule and bunnynet_pullzone_optimizer_class)
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
		p.pullzones[id].Lock()
		p.mu.Unlock()
	}
}

func (p *pullzoneMutex) Unlock(id int64) {
	p.mu.Lock()
	if _, ok := p.pullzones[id]; ok {
		p.pullzones[id].Unlock()
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
	minutes := int64(seconds / 60)
	remainder := math.Mod(float64(seconds), 60)

	return fmt.Sprintf("%02d:%02d", minutes, int64(remainder))
}

func generateMarkdownMapOptions[T comparable](m map[T]string) string {
	s := maps.Values(m)
	return generateMarkdownSliceOptions(s)
}

func generateMarkdownSliceOptions(options []string) string {
	// sorting a copy of the slice to avoid concurrency issues with validators
	s := slices.Clone(options)
	slices.Sort(s)
	return "Options: `" + strings.Join(s, "`, `") + "`"
}

func typeStringOrNull(value string) types.String {
	if value != "" {
		return types.StringValue(value)
	}

	return types.StringNull()
}
