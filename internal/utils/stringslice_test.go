package utils

import (
	"reflect"
	"testing"
)

func TestConvertCSVToStringSlice(t *testing.T) {
	type testCase struct {
		Expected []string
		Value    string
	}

	dataProvider := []testCase{
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a,b,c"},
		{[]string{"a", "b", "c"}, "a, b,\tc"},
		{[]string{"a", "b", "c"}, "\ta,b,c "},
		{[]string{"a", "b", "c"}, ",a,b,c,"},
		{[]string{"a", "b", "c"}, ",a,,,b,c,"},
	}

	for _, data := range dataProvider {
		result := ConvertCSVToStringSlice(data.Value)
		if !reflect.DeepEqual(result, data.Expected) {
			t.Errorf("Expected %v, got %v", data.Expected, result)
		}
	}
}
