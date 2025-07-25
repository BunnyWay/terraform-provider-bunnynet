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

func TestMiddlewareScript(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	originType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"middleware_script": tftypes.Number,
		},
	}

	routingFiltersElementType := tftypes.Set{
		ElementType: tftypes.String,
	}

	routingType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"filters": routingFiltersElementType,
		},
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"middleware_script": tftypes.NewValue(tftypes.Number, 123),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"middleware_script": tftypes.NewValue(tftypes.Number, 123),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "all"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"middleware_script": tftypes.NewValue(tftypes.Number, 123),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
	}

	configSchema := schema.Schema{
		Blocks: map[string]schema.Block{
			"origin": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"middleware_script": schema.Int64Attribute{},
				},
			},
			"routing": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"filters": schema.SetAttribute{
						ElementType: types.StringType,
					},
				},
			},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"origin":  originType,
			"routing": routingType,
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
		middlewareScriptValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
