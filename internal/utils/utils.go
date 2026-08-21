// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"strings"
)

func SliceDiff[T comparable](s1 []T, s2 []T) []T {
	diff := make([]T, 0)

	for _, v1 := range s1 {
		found := false
		for _, v2 := range s2 {
			if v1 == v2 {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, v1)
		}
	}

	return diff
}

func MapInvert[k comparable, v comparable](m map[k]v) map[v]k {
	result := make(map[v]k, len(m))
	for key, value := range m {
		result[value] = key
	}
	return result
}

func MapLowerKey[v comparable](m map[string]v) map[string]v {
	result := make(map[string]v, len(m))
	for key, value := range m {
		result[strings.ToLower(key)] = value
	}
	return result
}
