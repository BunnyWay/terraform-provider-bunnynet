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

func TestBotDetection(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	botDetectionType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"mode": tftypes.String,
		},
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":          tftypes.NewValue(tftypes.String, "Basic"),
				"bot_detection": tftypes.NewValue(botDetectionType, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Basic"),
				"bot_detection": tftypes.NewValue(botDetectionType, map[string]tftypes.Value{
					"mode": tftypes.NewValue(tftypes.String, "Log"),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":          tftypes.NewValue(tftypes.String, "Advanced"),
				"bot_detection": tftypes.NewValue(botDetectionType, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Advanced"),
				"bot_detection": tftypes.NewValue(botDetectionType, map[string]tftypes.Value{
					"mode": tftypes.NewValue(tftypes.String, "Log"),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":          tftypes.NewValue(tftypes.String, "Business"),
				"bot_detection": tftypes.NewValue(botDetectionType, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier": tftypes.NewValue(tftypes.String, "Business"),
				"bot_detection": tftypes.NewValue(botDetectionType, map[string]tftypes.Value{
					"mode": tftypes.NewValue(tftypes.String, "Log"),
				}),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tier": schema.StringAttribute{},
			"bot_detection": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"mode": types.StringType,
				},
			},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"tier":          tftypes.String,
			"bot_detection": botDetectionType,
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
		botDetectionValidator{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
