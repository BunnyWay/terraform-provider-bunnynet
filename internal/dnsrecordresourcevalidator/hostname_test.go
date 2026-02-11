package dnsrecordresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestHostname(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"type":  tftypes.NewValue(tftypes.String, "CNAME"),
				"value": tftypes.NewValue(tftypes.String, "server.example.com"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"type":  tftypes.NewValue(tftypes.String, "CNAME"),
				"value": tftypes.NewValue(tftypes.String, "server.example.com."),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"type":  tftypes.NewValue(tftypes.String, "TXT"),
				"value": tftypes.NewValue(tftypes.String, "some-value"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"type":  tftypes.NewValue(tftypes.String, "TXT"),
				"value": tftypes.NewValue(tftypes.String, "some-value."),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"type":  tftypes.NewValue(tftypes.String, "MX"),
				"value": tftypes.NewValue(tftypes.String, "."),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"type":  tftypes.NewValue(tftypes.String, "MX"),
				"value": tftypes.NewValue(tftypes.String, "mail.example.com."),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{},
			"value": schema.StringAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":  tftypes.String,
			"value": tftypes.String,
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
		hostnameValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
