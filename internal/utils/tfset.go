package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ConvertStringSliceToSetMust(values []string) types.Set {
	return types.SetValueMust(types.StringType, convertStringSliceToSetElements(values))
}

func ConvertStringSliceToSet(values []string) (types.Set, diag.Diagnostics) {
	return types.SetValue(types.StringType, convertStringSliceToSetElements(values))
}

func convertStringSliceToSetElements(values []string) []attr.Value {
	elements := make([]attr.Value, len(values))
	for i, v := range values {
		elements[i] = types.StringValue(v)
	}
	return elements
}
