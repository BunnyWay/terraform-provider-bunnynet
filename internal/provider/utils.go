package provider

import "golang.org/x/exp/constraints"

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
