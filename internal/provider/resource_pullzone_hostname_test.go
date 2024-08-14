// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configPullzoneHostnameTest = `
resource "bunnynet_pullzone" "test" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url = "https://192.0.2.1"
  }

  routing {
    tier = "Standard"
  }
}

resource "bunnynet_pullzone_hostname" "test" {
  pullzone = bunnynet_pullzone.test.id
  name = "test-acceptance-%s.terraform.internal"
  tls_enabled = %t
  force_ssl = %t
}
`

func TestAccPullzoneHostnameResource(t *testing.T) {
	resourceName := "bunnynet_pullzone_hostname.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneHostnameTest, testKey, testKey, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acceptance-%s.terraform.internal", testKey)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPullzoneHostnameImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config:      fmt.Sprintf(configPullzoneHostnameTest, testKey, testKey, false, true),
				ExpectError: regexp.MustCompile(`.*Attribute "tls_enabled" must also be set to true.`),
			},
			{
				// workaround for tests with ExpectError
				// {@see https://github.com/hashicorp/terraform-plugin-sdk/issues/609#issuecomment-1101251674}
				Config: fmt.Sprintf(configPullzoneHostnameTest, testKey, testKey, false, false),
			},
		},
	})
}

func testAccPullzoneHostnameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["pullzone"], rs.Primary.Attributes["name"]), nil
	}
}
