package streamlibraryresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestPremiumEncodingOutputCodecs(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "h264"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "vp9"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hevc"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "av1"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "h264"),
					tftypes.NewValue(tftypes.String, "vp9"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "h264"),
					tftypes.NewValue(tftypes.String, "hevc"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Free"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "h264"),
					tftypes.NewValue(tftypes.String, "av1"),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Premium"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "h264"),
					tftypes.NewValue(tftypes.String, "av1"),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Premium"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "vp9"),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"encoding_tier": tftypes.NewValue(tftypes.String, "Premium"),
				"output_codecs": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "h264"),
					tftypes.NewValue(tftypes.String, "vp9"),
					tftypes.NewValue(tftypes.String, "hevc"),
					tftypes.NewValue(tftypes.String, "av1"),
				}),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"encoding_tier": schema.StringAttribute{},
			"output_codecs": schema.SetAttribute{ElementType: types.StringType},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"encoding_tier": tftypes.String,
			"output_codecs": tftypes.Set{ElementType: tftypes.String},
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
		premiumEncodingValidator{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
