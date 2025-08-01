package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestAccessListSetUniqueValues(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		Values        []attr.Value
	}

	accessListType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":     types.Int64Type,
			"action": types.StringType,
		},
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			Values: []attr.Value{
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Value(1),
					"action": types.StringValue("Block"),
				}),
			},
		},
		{
			ExpectedError: false,
			Values: []attr.Value{
				types.ObjectUnknown(accessListType.AttrTypes),
			},
		},
		{
			ExpectedError: false,
			Values: []attr.Value{
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Unknown(),
					"action": types.StringValue("Block"),
				}),
			},
		},
		{
			ExpectedError: false,
			Values: []attr.Value{
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Unknown(),
					"action": types.StringValue("Block"),
				}),
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Unknown(),
					"action": types.StringValue("Log"),
				}),
			},
		},
		{
			ExpectedError: true,
			Values: []attr.Value{
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Value(1),
					"action": types.StringValue("Block"),
				}),
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Value(1),
					"action": types.StringValue("Log"),
				}),
			},
		},
		{
			ExpectedError: true,
			Values: []attr.Value{
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Value(1),
					"action": types.StringValue("Block"),
				}),
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Value(1),
					"action": types.StringValue("Log"),
				}),
				types.ObjectValueMust(accessListType.AttrTypes, map[string]attr.Value{
					"id":     types.Int64Value(1),
					"action": types.StringValue("Blog"),
				}),
			},
		},
	}

	for _, tc := range testCases {
		set, diags := types.SetValue(accessListType, tc.Values)
		if diags.HasError() {
			t.Error(diags)
			continue
		}

		request := validator.SetRequest{
			Path:        path.Root("access_list"),
			ConfigValue: set,
		}

		response := validator.SetResponse{}
		accessListSet{}.ValidateSet(context.Background(), request, &response)

		if tc.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !tc.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
