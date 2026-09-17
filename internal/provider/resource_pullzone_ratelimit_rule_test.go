// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

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
  tier     = "%s"

  ddos {
    level = "Asleep"
  }

  waf {
    enabled = false
  }
}

resource "bunnynet_pullzone_ratelimit_rule" "test" {
  pullzone        = bunnynet_pullzone_shield.test.pullzone
  name            = "WordPress Login"
  description     = "WordPress Login"
  transformations = ["LOWERCASE", "NORMALIZEPATH", "URLDECODE"]

  condition {
    variable = "REQUEST_URI"
    operator = "BEGINSWITH"
    value    = "/wp-login.php"
  }

  limit {
    requests = 2
    interval = %d # seconds
  }

  response {
    interval = %d # seconds
  }
}
`

func TestAccPullzoneRatelimitRuleResourceWithBasicPlan(t *testing.T) {
	shieldResource := "bunnynet_pullzone_shield.test"
	ratelimitResource := "bunnynet_pullzone_ratelimit_rule.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneRatelimitRuleTest, testKey, "Basic", 10, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(shieldResource, "tier", "Basic"),
					resource.TestCheckResourceAttr(ratelimitResource, "condition.0.negated", "false"),
					resource.TestCheckResourceAttr(ratelimitResource, "limit.counter_key", "IP"),
					resource.TestCheckResourceAttr(ratelimitResource, "limit.interval", "10"),
					resource.TestCheckResourceAttr(ratelimitResource, "response.action", "RateLimit"),
					resource.TestCheckResourceAttr(ratelimitResource, "response.interval", "60"),
				),
			},
			{
				ResourceName:      ratelimitResource,
				ImportState:       true,
				ImportStateIdFunc: testAccPullzoneRatelimitRuleImportStateIdFunc(ratelimitResource),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPullzoneRatelimitRuleResourceWithAdvancedPlan(t *testing.T) {
	shieldResource := "bunnynet_pullzone_shield.test"
	ratelimitResource := "bunnynet_pullzone_ratelimit_rule.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneRatelimitRuleTest, testKey, "Advanced", 60, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(shieldResource, "tier", "Advanced"),
					resource.TestCheckResourceAttr(ratelimitResource, "condition.0.negated", "false"),
					resource.TestCheckResourceAttr(ratelimitResource, "limit.counter_key", "IP"),
					resource.TestCheckResourceAttr(ratelimitResource, "limit.interval", "60"),
					resource.TestCheckResourceAttr(ratelimitResource, "response.action", "RateLimit"),
					resource.TestCheckResourceAttr(ratelimitResource, "response.interval", "300"),
				),
			},
			{
				ResourceName:      ratelimitResource,
				ImportState:       true,
				ImportStateIdFunc: testAccPullzoneRatelimitRuleImportStateIdFunc(ratelimitResource),
				ImportStateVerify: true,
			},
		},
	})
}

const configPullzoneRatelimitRuleWithActionTest = `
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
    enabled = false
  }
}

resource "bunnynet_pullzone_ratelimit_rule" "test" {
  pullzone        = bunnynet_pullzone_shield.test.pullzone
  name            = "WordPress Login"
  description     = "WordPress Login"
  transformations = ["LOWERCASE", "NORMALIZEPATH", "URLDECODE"]

  condition {
    variable = "REQUEST_METHOD"
    operator = "STREQ"
    value    = "get"
    negated  = "true"
  }

  limit {
    counter_key = "ASN"
    requests = 2
    interval = 10 # seconds
  }

  response {
    action = "%s"
    interval = 60 # seconds
  }
}
`

func TestAccPullzoneRatelimitRuleResourceWithChallengeAction(t *testing.T) {
	shieldResource := "bunnynet_pullzone_shield.test"
	ratelimitResource := "bunnynet_pullzone_ratelimit_rule.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneRatelimitRuleWithActionTest, testKey, "Challenge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(shieldResource, "tier", "Advanced"),
					resource.TestCheckResourceAttr(ratelimitResource, "condition.0.negated", "true"),
					resource.TestCheckResourceAttr(ratelimitResource, "limit.counter_key", "ASN"),
					resource.TestCheckResourceAttr(ratelimitResource, "response.action", "Challenge"),
					resource.TestCheckResourceAttr(ratelimitResource, "response.interval", "60"),
				),
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
