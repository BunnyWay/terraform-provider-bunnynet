// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configPullzoneOptimizerClassTest = `
resource "bunnynet_pullzone" "test" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url = "https://bunny.net"
  }

  routing {
    tier = "Standard"
  }
}

resource "bunnynet_pullzone_optimizer_class" "test" {
  pullzone = bunnynet_pullzone.test.id
  name     = "%s"
  quality  = 75
}
`

func TestAccPullzoneOptimizerClassResource(t *testing.T) {
	resourceName := "bunnynet_pullzone_optimizer_class.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneOptimizerClassTest, testKey, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", testKey),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccPullzoneOptimizerClassImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
			{
				// workaround for tests with ExpectError
				// {@see https://github.com/hashicorp/terraform-plugin-sdk/issues/609#issuecomment-1101251674}
				Config: fmt.Sprintf(configPullzoneOptimizerClassTest, testKey, testKey),
			},
		},
	})
}

func testAccPullzoneOptimizerClassImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["pullzone"], rs.Primary.Attributes["name"]), nil
	}
}
