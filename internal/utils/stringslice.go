package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

func ConvertSetToStringSlice(set types.Set) []string {
	elements := set.Elements()
	values := make([]string, len(elements))
	for i, el := range elements {
		values[i] = el.(types.String).ValueString()
	}
	return values
}

func ConvertCSVToStringSlice(csv string) []string {
	split := strings.Split(csv, ",")

	values := make([]string, 0, len(split))
	for _, v := range split {
		vTrimmed := strings.TrimSpace(v)
		if len(vTrimmed) == 0 {
			continue
		}

		values = append(values, vTrimmed)
	}

	return values
}
