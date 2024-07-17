// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configStorageZoneTest = `
resource "bunnynet_storage_zone" "test" {
  name = "test-acceptance-%s"
  region = "DE"
  zone_tier = "Standard"
}
`

func TestAccStorageZoneResource(t *testing.T) {
	resourceName := "bunnynet_storage_zone.test"
	testKey := generateRandomString(12)
	config := fmt.Sprintf(configStorageZoneTest, testKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acceptance-%s", testKey)),
					resource.TestCheckResourceAttr(resourceName, "region", "DE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStorageZoneImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStorageZoneImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}
