// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

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
				"tier":                  tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, nil),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, false),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, true),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, false),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, "<h1>Blocked</h1>"),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Basic"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, true),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, "<h1>Blocked</h1>"),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, nil),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, false),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, true),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, true),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, "<h1>Blocked</h1>"),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, false),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, "<h1>Blocked</h1>"),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, false),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, "<h1>Challenge</h1>"),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"tier":                  tftypes.NewValue(tftypes.String, "Advanced"),
				"whitelabel":            tftypes.NewValue(tftypes.Bool, false),
				"whitelabel_block":      tftypes.NewValue(tftypes.String, nil),
				"whitelabel_challenge":  tftypes.NewValue(tftypes.String, nil),
				"whitelabel_rate_limit": tftypes.NewValue(tftypes.String, "<h1>Rate limit</h1>"),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tier":                  schema.StringAttribute{},
			"whitelabel":            schema.BoolAttribute{},
			"whitelabel_block":      schema.StringAttribute{},
			"whitelabel_challenge":  schema.StringAttribute{},
			"whitelabel_rate_limit": schema.StringAttribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"tier":                  tftypes.String,
			"whitelabel":            tftypes.Bool,
			"whitelabel_block":      tftypes.String,
			"whitelabel_challenge":  tftypes.String,
			"whitelabel_rate_limit": tftypes.String,
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
