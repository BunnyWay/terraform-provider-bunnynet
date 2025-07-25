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

func TestOriginComputeScript(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	originType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type": tftypes.String,
			"url":  tftypes.String,
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
					"type": tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":  tftypes.NewValue(tftypes.String, "https://example.com.com"),
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
					"type": tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":  tftypes.NewValue(tftypes.String, "https://example.com.com"),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type": tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":  tftypes.NewValue(tftypes.String, "https://bunnycdn.com"),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type": tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":  tftypes.NewValue(tftypes.String, "https://bunnycdn.com"),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "all"),
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type": tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":  tftypes.NewValue(tftypes.String, nil),
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
					"type": tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":  tftypes.NewValue(tftypes.String, "https://example.com"),
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
					"type": tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":  tftypes.NewValue(tftypes.String, "https://bunnycdn.com"),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, nil),
				}),
			},
		},
	}

	configSchema := schema.Schema{
		Blocks: map[string]schema.Block{
			"origin": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{},
					"url":  schema.StringAttribute{},
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
		originComputeScriptValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
