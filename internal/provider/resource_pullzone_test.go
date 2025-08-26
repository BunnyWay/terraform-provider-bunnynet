// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configPullzoneTest = `
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
`

func TestAccPullzoneResource(t *testing.T) {
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneTest, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bunnynet_pullzone.test", "name", fmt.Sprintf("test-acceptance-%s", testKey)),
					resource.TestCheckResourceAttr("bunnynet_pullzone.test", "cache_expiration_time", "-1"),
				),
			},
		},
	})
}

const configPullzoneWithPermacacheTest = `
resource "bunnynet_storage_zone" "test" {
  name = "test-acceptance-%s"
  zone_tier = "Standard"
  region = "DE"
}

resource "bunnynet_pullzone" "test" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url = "https://bunny.net"
  }

  routing {
    tier = "Standard"
  }

  permacache_storagezone = bunnynet_storage_zone.test.id
}
`

func TestAccPullzoneWithPermacacheResource(t *testing.T) {
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneWithPermacacheTest, testKey, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bunnynet_storage_zone.test", "name", fmt.Sprintf("test-acceptance-%s", testKey)),
					resource.TestCheckResourceAttr("bunnynet_pullzone.test", "name", fmt.Sprintf("test-acceptance-%s", testKey)),
					resource.TestCheckResourceAttr("bunnynet_pullzone.test", "cache_expiration_time", "31919000"),
				),
			},
		},
	})
}
