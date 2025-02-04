package storagezoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestRegions(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		// standard
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Standard"),
				"region":              tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Standard"),
				"region":              tftypes.NewValue(tftypes.String, "SG"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			// region available only on Edge tier
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Standard"),
				"region":              tftypes.NewValue(tftypes.String, "JP"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Standard"),
				"region":    tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "LA"),
					tftypes.NewValue(tftypes.String, "SG"),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Standard"),
				"region":    tftypes.NewValue(tftypes.String, "SG"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "LA"),
					tftypes.NewValue(tftypes.String, "DE"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Standard"),
				"region":              tftypes.NewValue(tftypes.String, "ZZ"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Standard"),
				"region":    tftypes.NewValue(tftypes.String, "ZZ"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "LA"),
					tftypes.NewValue(tftypes.String, "SG"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Standard"),
				"region":    tftypes.NewValue(tftypes.String, "SG"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "LA"),
					tftypes.NewValue(tftypes.String, "SG"),
				}),
			},
		},
		{
			// region available only on Edge tier
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Standard"),
				"region":    tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "JP"),
				}),
			},
		},
		{
			// region repeated within replication_regions
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Standard"),
				"region":    tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "DE"),
					tftypes.NewValue(tftypes.String, "SG"),
				}),
			},
		},

		// edge
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Edge"),
				"region":              tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Edge"),
				"region":              tftypes.NewValue(tftypes.String, "ZZ"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier":           tftypes.NewValue(tftypes.String, "Edge"),
				"region":              tftypes.NewValue(tftypes.String, "SG"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Edge"),
				"region":    tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "LA"),
					tftypes.NewValue(tftypes.String, "SG"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Edge"),
				"region":    tftypes.NewValue(tftypes.String, "DE"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "ZZ"),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"zone_tier": tftypes.NewValue(tftypes.String, "Edge"),
				"region":    tftypes.NewValue(tftypes.String, "SG"),
				"replication_regions": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "DE"),
					tftypes.NewValue(tftypes.String, "LA"),
				}),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zone_tier":           schema.StringAttribute{},
			"region":              schema.StringAttribute{},
			"replication_regions": schema.SetAttribute{ElementType: types.StringType},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"zone_tier":           tftypes.String,
			"region":              tftypes.String,
			"replication_regions": tftypes.Set{ElementType: tftypes.String},
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
		regionValidator{}.ValidateResource(context.Background(), request, &response)

		if data.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !data.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
