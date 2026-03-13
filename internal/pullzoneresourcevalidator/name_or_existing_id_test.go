// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"testing"
)

func TestNameOrExistingId(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"name":        tftypes.NewValue(tftypes.String, "my-pullzone"),
				"existing_id": tftypes.NewValue(tftypes.Number, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"name":        tftypes.NewValue(tftypes.String, nil),
				"existing_id": tftypes.NewValue(tftypes.Number, 12345),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"name":        tftypes.NewValue(tftypes.String, "my-pullzone"),
				"existing_id": tftypes.NewValue(tftypes.Number, 12345),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"name":        tftypes.NewValue(tftypes.String, nil),
				"existing_id": tftypes.NewValue(tftypes.Number, nil),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{},
			"existing_id": schema.Int64Attribute{},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name":        tftypes.String,
			"existing_id": tftypes.Number,
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
		nameOrExistingIdValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
