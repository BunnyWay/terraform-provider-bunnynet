package utils

import "reflect"

func SliceToSet[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{}, len(slice))

	for _, item := range slice {
		set[item] = struct{}{}
	}

	return set
}

func SetEquals[T comparable](s1 map[T]struct{}, s2 map[T]struct{}) bool {
	return reflect.DeepEqual(s1, s2)
}

func SetToSlice[T comparable](set map[T]struct{}) []T {
	slice := make([]T, 0, len(set))

	for k := range set {
		slice = append(slice, k)
	}

	return slice
}
