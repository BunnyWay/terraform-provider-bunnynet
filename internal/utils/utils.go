// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package utils

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
