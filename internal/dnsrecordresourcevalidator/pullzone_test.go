package dnsrecordresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestPullzone(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"type":        tftypes.NewValue(tftypes.String, "TXT"),
				"pullzone_id": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"type":        tftypes.NewValue(tftypes.String, "TXT"),
				"pullzone_id": tftypes.NewValue(tftypes.Number, 1234),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"type":        tftypes.NewValue(tftypes.String, "PullZone"),
				"pullzone_id": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"type":        tftypes.NewValue(tftypes.String, "PullZone"),
				"pullzone_id": tftypes.NewValue(tftypes.Number, 1234),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":        schema.StringAttribute{},
			"pullzone_id": schema.Int64Attribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":        tftypes.String,
			"pullzone_id": tftypes.Number,
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
		pullzoneIdValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
