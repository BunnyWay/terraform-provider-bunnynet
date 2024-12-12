package pullzoneedgeruleresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestAction(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"action":            tftypes.NewValue(tftypes.String, "Redirect"),
				"action_parameter1": tftypes.NewValue(tftypes.String, "https://bunny.net"),
				"action_parameter2": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"action":            tftypes.NewValue(tftypes.String, "Redirect"),
				"action_parameter1": tftypes.NewValue(tftypes.String, "https://bunny.net"),
				"action_parameter2": tftypes.NewValue(tftypes.String, "301"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"action":            tftypes.NewValue(tftypes.String, "Redirect"),
				"action_parameter1": tftypes.NewValue(tftypes.String, "https://bunny.net"),
				"action_parameter2": tftypes.NewValue(tftypes.String, "400"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"action":            tftypes.NewValue(tftypes.String, "Redirect"),
				"action_parameter1": tftypes.NewValue(tftypes.String, "https://bunny.net"),
				"action_parameter2": tftypes.NewValue(tftypes.String, "abc"),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action":            schema.StringAttribute{},
			"action_parameter1": schema.StringAttribute{},
			"action_parameter2": schema.StringAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"action":            tftypes.String,
			"action_parameter1": tftypes.String,
			"action_parameter2": tftypes.String,
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
		actionParameters{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
