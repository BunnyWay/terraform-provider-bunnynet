package pullzoneshieldresourcevalidator

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

func TestRealtimeThreatIntelligence(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	wafObjType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"realtime_threat_intelligence": tftypes.Bool,
		},
	}

	testCases := []testCase{
		// @TODO can't test with realtime_threat_intelligence missing, as the framework does not support OptionalAttributes
		// panic: can't create a tftypes.Value of type tftypes.Object["realtime_threat_intelligence":tftypes.Bool], required attribute "realtime_threat_intelligence" not set
		// panic: Objects with OptionalAttributes cannot be used.
		//{
		//	ExpectedError: false,
		//	PlanValues: map[string]tftypes.Value{
		//		"tier": tftypes.NewValue(tftypes.String, "Basic"),
		//		"waf":  tftypes.NewValue(wafObjType, map[string]tftypes.Value{}),
		//	},
		//},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Basic"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Advanced"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Business"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Enterprise"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Basic"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, false),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Advanced"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, false),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Business"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, false),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Enterprise"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, false),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Basic"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, true),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Advanced"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, true),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Business"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, true),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Enterprise"),
				"waf": tftypes.NewValue(wafObjType, map[string]tftypes.Value{
					"realtime_threat_intelligence": tftypes.NewValue(tftypes.Bool, true),
				}),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tier": schema.StringAttribute{},
			"waf": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"realtime_threat_intelligence": types.BoolType,
				},
			},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"tier": tftypes.String,
			"waf":  wafObjType,
		},
	}

	for _, data := range testCases {
		request := resource.ValidateConfigRequest{
			Config: tfsdk.Config{
				Schema: configSchema,
				Raw:    tftypes.NewValue(configTypes, data.PlanValues),
			},
		}

		response := resource.ValidateConfigResponse{}
		realtimeThreatIntelligenceValidator{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
