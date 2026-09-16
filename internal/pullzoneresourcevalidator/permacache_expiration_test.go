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

func TestPermacacheCacheExpirationTime(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, nil),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, -1),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, 3600),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, 3600),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 0),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, nil),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, 3600),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, DefaultCacheExpirationTimeForPermacache),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 123),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, DefaultCacheExpirationTimeForPermacache),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, nil),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, 3600),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"cache_expiration_time":  tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				"permacache_storagezone": tftypes.NewValue(tftypes.Number, 0),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cache_expiration_time":  schema.Int64Attribute{},
			"permacache_storagezone": schema.Int64Attribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"cache_expiration_time":  tftypes.Number,
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
			permacacheCacheExpirationTimeValidator{}.ValidateResource(context.Background(), request, &response)

			if tc.ExpectedError && !response.Diagnostics.HasError() {
				t.Error("expected error, got none")
			}

			if !tc.ExpectedError && response.Diagnostics.HasError() {
				t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
			}
		})
	}
}
