package computecontainerappresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestEndpointNameShouldBeUnique(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		Values        [][]map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "anycast"),
						"type": tftypes.NewValue(tftypes.String, "anycast"),
					},
				},
			},
		},
		{
			ExpectedError: false,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
				{
					{
						"name": tftypes.NewValue(tftypes.String, "anycast"),
						"type": tftypes.NewValue(tftypes.String, "anycast"),
					},
				},
			},
		},
		{
			ExpectedError: false,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "http2"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
			},
		},
		{
			ExpectedError: false,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http2"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
			},
		},
		{
			ExpectedError: true,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
			},
		},
		{
			ExpectedError: true,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "anycast"),
						"type": tftypes.NewValue(tftypes.String, "anycast"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
			},
		},
		{
			ExpectedError: true,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "anycast"),
						"type": tftypes.NewValue(tftypes.String, "anycast"),
					},
				},
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
			},
		},
		{
			ExpectedError: true,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "anycast"),
					},
				},
			},
		},
		{
			ExpectedError: true,
			Values: [][]map[string]tftypes.Value{
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "cdn"),
					},
				},
				{
					{
						"name": tftypes.NewValue(tftypes.String, "http"),
						"type": tftypes.NewValue(tftypes.String, "anycast"),
					},
				},
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"container": schema.ListAttribute{ElementType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"endpoint": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name": types.StringType,
								"type": types.StringType,
							},
						},
					},
				},
			}},
		},
	}

	endpointType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
			"type": tftypes.String,
		},
	}

	endpointListType := tftypes.List{
		ElementType: endpointType,
	}

	containerType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"endpoint": endpointListType,
		},
	}

	containerListType := tftypes.List{
		ElementType: containerType,
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"container": containerListType,
		},
	}

	for iTestCase, testCase := range testCases {
		containers := make([]tftypes.Value, len(testCase.Values))

		for j, endpointArr := range testCase.Values {
			endpoints := make([]tftypes.Value, len(endpointArr))
			for i, endpointAttrs := range endpointArr {
				endpoints[i] = tftypes.NewValue(endpointType, endpointAttrs)
			}

			containers[j] = tftypes.NewValue(configTypes.AttributeTypes["container"].(tftypes.List).ElementType, map[string]tftypes.Value{
				"endpoint": tftypes.NewValue(configTypes.AttributeTypes["container"].(tftypes.List).ElementType.(tftypes.Object).AttributeTypes["endpoint"], endpoints),
			})
		}

		request := resource.ValidateConfigRequest{
			Config: tfsdk.Config{
				Schema: configSchema,
				Raw: tftypes.NewValue(configTypes, map[string]tftypes.Value{
					"container": tftypes.NewValue(configTypes.AttributeTypes["container"], containers),
				}),
			},
		}

		response := resource.ValidateConfigResponse{}
		containerEndpointNameValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Errorf("#%d: expected error, got none", iTestCase)
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("#%d: expected no errors, got %s", iTestCase, response.Diagnostics.Errors())
		}
	}
}
