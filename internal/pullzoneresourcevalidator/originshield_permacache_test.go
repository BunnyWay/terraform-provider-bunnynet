// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"testing"
)

func TestOriginShieldPermacache(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, nil),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, true),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, true),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 0),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, false),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, nil),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, true),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"originshield_enabled":   tftypes.NewValue(tftypes.Bool, true),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"originshield_enabled":   schema.BoolAttribute{},
			"permacache_storagezone": schema.Int64Attribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"originshield_enabled":   tftypes.Bool,
			"permacache_storagezone": tftypes.Number,
		},
	}

	for n, tc := range testCases {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			request := resource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Schema: configSchema,
					Raw:    tftypes.NewValue(configTypes, tc.PlanValues),
				},
			}

			response := resource.ValidateConfigResponse{}
			OriginShieldPermacache().ValidateResource(context.Background(), request, &response)

			if tc.ExpectedError && !response.Diagnostics.HasError() {
				t.Error("expected error, got none")
			}

			if !tc.ExpectedError && response.Diagnostics.HasError() {
				t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
			}
		})
	}
}
