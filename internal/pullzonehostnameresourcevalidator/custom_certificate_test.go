package pullzonehostnameresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"testing"
)

func TestCustomCertificate(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tls_enabled": tftypes.NewValue(tftypes.Bool, nil),
				"certificate": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tls_enabled": tftypes.NewValue(tftypes.Bool, nil),
				"certificate": tftypes.NewValue(tftypes.String, "test"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tls_enabled": tftypes.NewValue(tftypes.Bool, false),
				"certificate": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tls_enabled": tftypes.NewValue(tftypes.Bool, false),
				"certificate": tftypes.NewValue(tftypes.String, "test"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tls_enabled": tftypes.NewValue(tftypes.Bool, true),
				"certificate": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tls_enabled": tftypes.NewValue(tftypes.Bool, true),
				"certificate": tftypes.NewValue(tftypes.String, "test"),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tls_enabled": schema.BoolAttribute{},
			"certificate": schema.StringAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"tls_enabled": tftypes.Bool,
			"certificate": tftypes.String,
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
		customCertificateValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
