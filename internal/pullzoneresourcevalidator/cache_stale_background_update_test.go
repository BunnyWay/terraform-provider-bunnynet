package pullzoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestCacheStaleBackgroundUpdate(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale":           tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
				"use_background_update": tftypes.NewValue(tftypes.Bool, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale":           tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{}),
				"use_background_update": tftypes.NewValue(tftypes.Bool, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "offline"),
					tftypes.NewValue(tftypes.String, "updating"),
				}),
				"use_background_update": tftypes.NewValue(tftypes.Bool, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale":           tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
				"use_background_update": tftypes.NewValue(tftypes.Bool, false),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale":           tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{}),
				"use_background_update": tftypes.NewValue(tftypes.Bool, false),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"cache_stale": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "offline"),
					tftypes.NewValue(tftypes.String, "updating"),
				}),
				"use_background_update": tftypes.NewValue(tftypes.Bool, false),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale":           tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
				"use_background_update": tftypes.NewValue(tftypes.Bool, true),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale":           tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{}),
				"use_background_update": tftypes.NewValue(tftypes.Bool, true),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_stale": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "offline"),
					tftypes.NewValue(tftypes.String, "updating"),
				}),
				"use_background_update": tftypes.NewValue(tftypes.Bool, true),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cache_stale":           schema.SetAttribute{ElementType: types.StringType},
			"use_background_update": schema.BoolAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"cache_stale":           tftypes.Set{ElementType: tftypes.String},
			"use_background_update": tftypes.Bool,
		},
	}

	for _, testCase := range testCases {
		request := resource.ValidateConfigRequest{
			Config: tfsdk.Config{
				Schema: configSchema,
				Raw:    tftypes.NewValue(configTypes, testCase.PlanValues),
			},
		}

		response := resource.ValidateConfigResponse{}
		cacheStaleBackgroundUpdateValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s: %+v", response.Diagnostics.Errors(), testCase)
		}
	}
}
