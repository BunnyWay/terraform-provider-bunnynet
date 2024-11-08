// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configPullzoneEdgeruleTest = `
resource "bunnynet_pullzone" "test" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url = "https://bunny.net/"
  }

  routing {
    tier = "Standard"
  }
}

resource "bunnynet_pullzone_edgerule" "test" {
  enabled     = true
  pullzone    = bunnynet_pullzone.test.id
  action      = "BlockRequest"
  description = "Block access to admin"

  match_type = "MatchAny"
  triggers = [
    {
      type       = "%s"
      match_type = "%s"
      patterns   = ["*/wp-admin/*"]
      parameter1 = null
      parameter2 = null
    }
  ]
}
`

func TestAccPullzoneEdgeruleResource(t *testing.T) {
	resourceName := "bunnynet_pullzone_edgerule.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneEdgeruleTest, testKey, "Url", "MatchAny"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "BlockRequest"),
				),
			},
			{
				Config:      fmt.Sprintf(configPullzoneEdgeruleTest, testKey, "Invalid", "MatchAny"),
				ExpectError: regexp.MustCompile(`Error: Trigger type must be valid`),
			},
			{
				Config:      fmt.Sprintf(configPullzoneEdgeruleTest, testKey, "Url", "Invalid"),
				ExpectError: regexp.MustCompile(`Error: Trigger match_type must be valid`),
			},
			{
				// workaround for tests with ExpectError
				// {@see https://github.com/hashicorp/terraform-plugin-sdk/issues/609#issuecomment-1101251674}
				Config: fmt.Sprintf(configPullzoneEdgeruleTest, testKey, "Url", "MatchAny"),
			},
		},
	})
}
