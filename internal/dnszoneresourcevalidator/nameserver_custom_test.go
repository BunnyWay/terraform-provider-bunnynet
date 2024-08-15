package dnszoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"testing"
)

func TestCustomNameserverValidateResource(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	testCases := []testCase{
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, nil),
				"nameserver1":       tftypes.NewValue(tftypes.String, nil),
				"nameserver2":       tftypes.NewValue(tftypes.String, nil),
				"soa_email":         tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, false),
				"nameserver1":       tftypes.NewValue(tftypes.String, "kiki.bunny.net"),
				"nameserver2":       tftypes.NewValue(tftypes.String, "coco.bunny.net"),
				"soa_email":         tftypes.NewValue(tftypes.String, "hostmaster@bunny.net"),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, true),
				"nameserver1":       tftypes.NewValue(tftypes.String, "ns1.example.org"),
				"nameserver2":       tftypes.NewValue(tftypes.String, "ns2.example.org"),
				"soa_email":         tftypes.NewValue(tftypes.String, "hostmaster@example.org"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, false),
				"nameserver1":       tftypes.NewValue(tftypes.String, nil),
				"nameserver2":       tftypes.NewValue(tftypes.String, nil),
				"soa_email":         tftypes.NewValue(tftypes.String, "hostmaster@example.org"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, false),
				"nameserver1":       tftypes.NewValue(tftypes.String, "ns1.example.org"),
				"nameserver2":       tftypes.NewValue(tftypes.String, "ns2.example.org"),
				"soa_email":         tftypes.NewValue(tftypes.String, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, true),
				"nameserver1":       tftypes.NewValue(tftypes.String, "ns1.example.org"),
				"nameserver2":       tftypes.NewValue(tftypes.String, "ns2.example.org"),
				"soa_email":         tftypes.NewValue(tftypes.String, "hostmaster@bunny.net"),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"nameserver_custom": tftypes.NewValue(tftypes.Bool, false),
				"nameserver1":       tftypes.NewValue(tftypes.String, "ns1.example.org"),
				"nameserver2":       tftypes.NewValue(tftypes.String, "ns2.example.org"),
				"soa_email":         tftypes.NewValue(tftypes.String, "hostmaster@example.org"),
			},
		},
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"nameserver_custom": schema.BoolAttribute{
				Default: booldefault.StaticBool(false),
			},
			"nameserver1": schema.StringAttribute{
				Default: stringdefault.StaticString("kiki.bunny.net"),
			},
			"nameserver2": schema.StringAttribute{
				Default: stringdefault.StaticString("coco.bunny.net"),
			},
			"soa_email": schema.StringAttribute{
				Default: stringdefault.StaticString("hostmaster@bunny.net"),
			},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"nameserver_custom": tftypes.Bool,
			"nameserver1":       tftypes.String,
			"nameserver2":       tftypes.String,
			"soa_email":         tftypes.String,
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
		customNameserverValidator{}.ValidateResource(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
