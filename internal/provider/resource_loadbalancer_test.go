// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configLoadbalancerTest = `
resource "bunnynet_loadbalancer" "test" {
  name            = "test-acceptance-%s"
  method          = "Latency"
  sticky_sessions = "Cookie"
}

resource "bunnynet_loadbalancer_group" "EU" {
  loadbalancer        = bunnynet_loadbalancer.test.id
  name                = "EU"
  method              = "Latency"
  priority            = 1
  healthcheck_enabled = true
  healthcheck_path    = "/health"
}

resource "bunnynet_loadbalancer_origin" "DK" {
  loadbalancer = bunnynet_loadbalancer.test.id
  group        = bunnynet_loadbalancer_group.EU.id
  name         = "DK"
  type         = "OriginUrl"
  url          = "https://origin-dk.example.net"
}

resource "bunnynet_loadbalancer_origin" "IT" {
  loadbalancer = bunnynet_loadbalancer.test.id
  group        = bunnynet_loadbalancer_group.EU.id
  name         = "IT"
  type         = "OriginUrl"
  url          = "https://origin-it.example.net"
}
`

func TestAccLoadbalancerResource(t *testing.T) {
	testKey := generateRandomString(4)
	lbName := fmt.Sprintf("test-acceptance-%s", testKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configLoadbalancerTest, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bunnynet_loadbalancer.test", "name", lbName),
					resource.TestCheckResourceAttr("bunnynet_loadbalancer_group.EU", "name", "EU"),
					resource.TestCheckResourceAttr("bunnynet_loadbalancer_origin.DK", "name", "DK"),
					resource.TestCheckResourceAttr("bunnynet_loadbalancer_origin.IT", "name", "IT"),
				),
			},
			{
				ResourceName:      "bunnynet_loadbalancer.test",
				ImportState:       true,
				ImportStateIdFunc: testAccLoadbalancerImportStateIdFunc("bunnynet_loadbalancer.test"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "bunnynet_loadbalancer_group.EU",
				ImportState:       true,
				ImportStateIdFunc: testAccLoadbalancerGroupImportStateIdFunc("bunnynet_loadbalancer_group.EU"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "bunnynet_loadbalancer_origin.DK",
				ImportState:       true,
				ImportStateIdFunc: testAccLoadbalancerOriginImportStateIdFunc("bunnynet_loadbalancer_origin.DK"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccLoadbalancerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}

func testAccLoadbalancerGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["loadbalancer"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccLoadbalancerOriginImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		var groupName string
		for _, r := range s.RootModule().Resources {
			if r.Primary.Attributes["id"] == rs.Primary.Attributes["group"] {
				groupName = r.Primary.Attributes["name"]
			}
		}

		if groupName == "" {
			return "", fmt.Errorf("Could not find group %s", rs.Primary.Attributes["group"])
		}

		return fmt.Sprintf("%s|%s|%s", rs.Primary.Attributes["loadbalancer"], groupName, rs.Primary.Attributes["name"]), nil
	}
}
