package api

import (
	"golang.org/x/exp/slices"
	"testing"
)

func TestSliceDiff(t *testing.T) {
	type dataType struct {
		Expected []string
		SliceA   []string
		SliceB   []string
	}

	dataProvider := []dataType{
		{[]string{}, []string{"A"}, []string{"A"}},
		{[]string{"A"}, []string{"A"}, []string{}},
		{[]string{}, []string{}, []string{"A"}},
		{[]string{"B"}, []string{"A", "A", "B"}, []string{"A"}},
		{[]string{"B"}, []string{"A", "B"}, []string{"A", "A"}},
	}

	for _, data := range dataProvider {
		diff := sliceDiff(data.SliceA, data.SliceB)

		if !slices.Equal(diff, data.Expected) {
			t.Errorf("Expected diff to be %v, got %v", data.Expected, diff)
		}
	}
}
