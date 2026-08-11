package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestUploadScanning(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, nil),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Disable"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, nil),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Disable"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Disable"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Log"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Disable"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Block"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Log"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Disable"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Basic"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Block"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Disable"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Advanced"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Log"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Disable"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                      tftypes.NewValue(tftypes.String, "Business"),
				"upload_scanning_antivirus": tftypes.NewValue(tftypes.String, "Log"),
				"upload_scanning_csam":      tftypes.NewValue(tftypes.String, "Disable"),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tier":                      schema.StringAttribute{},
			"upload_scanning_antivirus": schema.StringAttribute{},
			"upload_scanning_csam":      schema.StringAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"tier":                      tftypes.String,
			"upload_scanning_antivirus": tftypes.String,
			"upload_scanning_csam":      tftypes.String,
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
		uploadScanningValidator{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
