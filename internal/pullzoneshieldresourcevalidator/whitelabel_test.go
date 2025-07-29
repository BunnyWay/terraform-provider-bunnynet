package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestWhitelabel(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":       tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel": tftypes.NewValue(tftypes.Bool, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":       tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel": tftypes.NewValue(tftypes.Bool, false),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":       tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel": tftypes.NewValue(tftypes.Bool, true),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":       tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel": tftypes.NewValue(tftypes.Bool, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":       tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel": tftypes.NewValue(tftypes.Bool, false),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":       tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel": tftypes.NewValue(tftypes.Bool, true),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tier":       schema.StringAttribute{},
			"whitelabel": schema.BoolAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"tier":       tftypes.String,
			"whitelabel": tftypes.Bool,
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
		whitelabelValidator{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
