// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/resourcestateupgrader"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestPullzoneRatelimitRuleModelConversion(t *testing.T) {
	conditions := types.ListValueMust(pullzoneRatelimitConditionType, []attr.Value{
		types.ObjectValueMust(pullzoneRatelimitConditionType.AttrTypes, map[string]attr.Value{
			"variable":       types.StringValue("REQUEST_URI"),
			"variable_value": types.StringNull(),
			"operator":       types.StringValue("BEGINSWITH"),
			"value":          types.StringValue("/wp-login.php"),
			"negated":        types.BoolValue(false),
		}),
		types.ObjectValueMust(pullzoneRatelimitConditionType.AttrTypes, map[string]attr.Value{
			"variable":       types.StringValue("REQUEST_COOKIES_NAMES"),
			"variable_value": types.StringNull(),
			"operator":       types.StringValue("CONTAINS"),
			"value":          types.StringValue("session"),
			"negated":        types.BoolValue(true),
		}),
	})

	model := PullzoneRatelimitRuleResourceModel{
		Id:              types.Int64Value(123),
		PullzoneId:      types.Int64Value(456),
		Name:            types.StringValue("WordPress Login"),
		Description:     types.StringValue("Login attempts without a session cookie"),
		Conditions:      conditions,
		Transformations: types.SetNull(types.StringType),
		Limit: types.ObjectValueMust(pullzoneRatelimitRuleLimitType, map[string]attr.Value{
			"requests":    types.Int64Value(5),
			"interval":    types.Int64Value(60),
			"counter_key": types.StringValue("IP"),
		}),
		Response: types.ObjectValueMust(pullzoneRatelimitRuleResponseType, map[string]attr.Value{
			"interval": types.Int64Value(30),
			"action":   types.StringValue("Log"),
		}),
	}

	providerResource := PullzoneRatelimitRuleResource{}
	apiRule := providerResource.convertModelToApi(context.Background(), model)

	if apiRule.RuleConfiguration.ActionType != 2 {
		t.Fatalf("expected Log action type 2, got %d", apiRule.RuleConfiguration.ActionType)
	}
	if apiRule.RuleConfiguration.CounterKeyType != 0 {
		t.Fatalf("expected IP counter key type 0, got %d", apiRule.RuleConfiguration.CounterKeyType)
	}
	if apiRule.RuleConfiguration.IsNegated {
		t.Fatal("expected primary condition not to be negated")
	}
	if len(apiRule.RuleConfiguration.ChainedRules) != 1 || !apiRule.RuleConfiguration.ChainedRules[0].IsNegated {
		t.Fatal("expected chained cookie condition to be negated")
	}

	roundTripModel, diags := providerResource.convertApiToModel(context.Background(), apiRule)
	if diags.HasError() {
		t.Fatalf("unexpected conversion diagnostics: %v", diags)
	}
	if roundTripModel.Response.Attributes()["action"].(types.String).ValueString() != "Log" {
		t.Fatalf("expected Log action, got %q", roundTripModel.Response.Attributes()["action"].(types.String).ValueString())
	}
	if roundTripModel.Limit.Attributes()["counter_key"].(types.String).ValueString() != "IP" {
		t.Fatalf("expected IP counter key, got %q", roundTripModel.Limit.Attributes()["counter_key"].(types.String).ValueString())
	}

	negatedByVariable := make(map[string]bool)
	for _, condition := range roundTripModel.Conditions.Elements() {
		values := condition.(types.Object).Attributes()
		negatedByVariable[values["variable"].(types.String).ValueString()] = values["negated"].(types.Bool).ValueBool()
	}
	if negatedByVariable["REQUEST_URI"] {
		t.Fatal("expected REQUEST_URI condition not to be negated after round trip")
	}
	if !negatedByVariable["REQUEST_COOKIES_NAMES"] {
		t.Fatal("expected REQUEST_COOKIES_NAMES condition to be negated after round trip")
	}
}

func TestPullzoneRatelimitRuleApiConversionChallenge(t *testing.T) {
	providerResource := PullzoneRatelimitRuleResource{}
	model, diags := providerResource.convertApiToModel(context.Background(), api.PullzoneRatelimitRule{
		RuleConfiguration: api.PullzoneRatelimitRuleConfiguration{
			ActionType:     3,
			CounterKeyType: 7,
			VariableTypes:  map[string]string{"REQUEST_URI": ""},
			OperatorType:   0,
			Value:          "/wp-login.php",
			RequestCount:   5,
			Timeframe:      60,
			BlockTime:      30,
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected conversion diagnostics: %v", diags)
	}
	if model.Response.Attributes()["action"].(types.String).ValueString() != "Challenge" {
		t.Fatalf("expected Challenge action, got %q", model.Response.Attributes()["action"].(types.String).ValueString())
	}
	if model.Limit.Attributes()["counter_key"].(types.String).ValueString() != "IP_JA4" {
		t.Fatalf("expected IP_JA4 counter key, got %q", model.Limit.Attributes()["counter_key"].(types.String).ValueString())
	}

	apiRule := providerResource.convertModelToApi(context.Background(), model)
	if apiRule.RuleConfiguration.ActionType != 3 {
		t.Fatalf("expected Challenge action type 3, got %d", apiRule.RuleConfiguration.ActionType)
	}
	if apiRule.RuleConfiguration.CounterKeyType != 7 {
		t.Fatalf("expected IP_JA4 counter key type 7, got %d", apiRule.RuleConfiguration.CounterKeyType)
	}
}

func TestPullzoneRatelimitRuleStateUpgradeV0(t *testing.T) {
	ctx := context.Background()

	stateV0 := `{
		"id": 123,
		"pullzone": 456,
		"name": "WordPress Login",
		"description": "WordPress Login",
		"condition": {
			"operator": "BEGINSWITH",
			"value": "/wp-login.php",
			"variable": "REQUEST_URI",
			"variable_value": "",
			"transformations": ["LOWERCASE", "NORMALIZEPATH"]
		},
		"limit": {"requests": 2, "interval": 10},
		"response": {"interval": 3600}
	}`

	schemaResp := &fwresource.SchemaResponse{}
	(&PullzoneRatelimitRuleResource{}).Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}

	upgradeResp := &fwresource.UpgradeStateResponse{}
	resourcestateupgrader.PullzoneRatelimitRuleV0(ctx, fwresource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(stateV0)},
	}, upgradeResp)

	if upgradeResp.Diagnostics.HasError() {
		t.Fatalf("unexpected upgrade diagnostics: %v", upgradeResp.Diagnostics)
	}
	if upgradeResp.DynamicValue == nil {
		t.Fatal("expected an upgraded state")
	}

	// Terraform decodes the upgraded state against the current schema, so every
	// attribute added to the schema must also be present here.
	upgraded, err := upgradeResp.DynamicValue.Unmarshal(schemaResp.Schema.Type().TerraformType(ctx))
	if err != nil {
		t.Fatalf("upgraded state does not match the current schema: %s", err)
	}

	var state map[string]tftypes.Value
	if err := upgraded.As(&state); err != nil {
		t.Fatalf("failed to convert upgraded state: %s", err)
	}

	var conditions []tftypes.Value
	if err := state["condition"].As(&conditions); err != nil {
		t.Fatalf("failed to convert conditions: %s", err)
	}
	if len(conditions) != 1 {
		t.Fatalf("expected a single condition, got %d", len(conditions))
	}

	var condition map[string]tftypes.Value
	if err := conditions[0].As(&condition); err != nil {
		t.Fatalf("failed to convert condition: %s", err)
	}

	var negated bool
	if err := condition["negated"].As(&negated); err != nil {
		t.Fatalf("failed to convert negated: %s", err)
	}
	if negated {
		t.Fatal("expected the upgraded condition not to be negated")
	}

	var limit map[string]tftypes.Value
	if err := state["limit"].As(&limit); err != nil {
		t.Fatalf("failed to convert limit: %s", err)
	}

	var counterKey string
	if err := limit["counter_key"].As(&counterKey); err != nil {
		t.Fatalf("failed to convert counter_key: %s", err)
	}
	if counterKey != "IP" {
		t.Fatalf("expected IP counter key, got %q", counterKey)
	}

	var response map[string]tftypes.Value
	if err := state["response"].As(&response); err != nil {
		t.Fatalf("failed to convert response: %s", err)
	}

	var action string
	if err := response["action"].As(&action); err != nil {
		t.Fatalf("failed to convert action: %s", err)
	}
	if action != "RateLimit" {
		t.Fatalf("expected RateLimit action, got %q", action)
	}
}

// rate limit rules need a paid Shield tier, Basic returns plan_limit_restriction.rate_limit
const configPullzoneRatelimitRuleTest = `
resource "bunnynet_pullzone" "test" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url  = "https://bunny.net"
  }

  routing {
    tier = "Standard"
  }
}

resource "bunnynet_pullzone_shield" "test" {
  pullzone = bunnynet_pullzone.test.id
  tier     = "Advanced"

  ddos {
    level = "Asleep"
  }

  waf {
    enabled = true
    mode    = "Log"
  }
}

resource "bunnynet_pullzone_ratelimit_rule" "test" {
  depends_on  = [bunnynet_pullzone_shield.test]
  pullzone    = bunnynet_pullzone.test.id
  name        = "WordPress Login"
  description = "Login attempts without a session cookie"

  condition {
    variable = "REQUEST_COOKIES_NAMES"
    operator = "CONTAINS"
    value    = "session"
    negated  = %s
  }

  condition {
    variable = "REQUEST_URI"
    operator = "BEGINSWITH"
    value    = "/wp-login.php"
  }

  limit {
    requests    = 5
    interval    = 60
    counter_key = "%s"
  }

  response {
    interval = 30
    action   = "%s"
  }
}
`

// same rule without action, counter_key and negated, as older configurations have it
const configPullzoneRatelimitRuleTestLegacy = `
resource "bunnynet_pullzone" "test" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url  = "https://bunny.net"
  }

  routing {
    tier = "Standard"
  }
}

resource "bunnynet_pullzone_shield" "test" {
  pullzone = bunnynet_pullzone.test.id
  tier     = "Advanced"

  ddos {
    level = "Asleep"
  }

  waf {
    enabled = true
    mode    = "Log"
  }
}

resource "bunnynet_pullzone_ratelimit_rule" "test" {
  depends_on  = [bunnynet_pullzone_shield.test]
  pullzone    = bunnynet_pullzone.test.id
  name        = "WordPress Login"
  description = "Login attempts without a session cookie"

  condition {
    variable = "REQUEST_COOKIES_NAMES"
    operator = "CONTAINS"
    value    = "session"
  }

  condition {
    variable = "REQUEST_URI"
    operator = "BEGINSWITH"
    value    = "/wp-login.php"
  }

  limit {
    requests = 5
    interval = 60
  }

  response {
    interval = 30
  }
}
`

func TestAccPullzoneRatelimitRuleResource(t *testing.T) {
	resourceName := "bunnynet_pullzone_ratelimit_rule.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneRatelimitRuleTest, testKey, "true", "IP", "Log"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "response.action", "Log"),
					resource.TestCheckResourceAttr(resourceName, "response.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "limit.counter_key", "IP"),
					resource.TestCheckResourceAttr(resourceName, "limit.requests", "5"),
					resource.TestCheckResourceAttr(resourceName, "limit.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.negated", "true"),
					resource.TestCheckResourceAttr(resourceName, "condition.1.negated", "false"),
				),
			},
			{
				Config: fmt.Sprintf(configPullzoneRatelimitRuleTest, testKey, "false", "Country", "Challenge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "response.action", "Challenge"),
					resource.TestCheckResourceAttr(resourceName, "limit.counter_key", "Country"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.negated", "false"),
					resource.TestCheckResourceAttr(resourceName, "condition.1.negated", "false"),
				),
			},
			{
				Config: fmt.Sprintf(configPullzoneRatelimitRuleTestLegacy, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "response.action", "RateLimit"),
					resource.TestCheckResourceAttr(resourceName, "limit.counter_key", "IP"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.negated", "false"),
					resource.TestCheckResourceAttr(resourceName, "condition.1.negated", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPullzoneRatelimitRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPullzoneRatelimitRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["pullzone"], rs.Primary.Attributes["id"]), nil
	}
}
