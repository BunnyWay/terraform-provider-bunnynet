// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
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

const configPullzoneExistingIdTest = `
data "bunnynet_dns_zone" "domain" {
  domain = "terraform.internal"
}

resource "bunnynet_dns_record" "accelerated" {
  zone        = data.bunnynet_dns_zone.domain.id
  name        = "test-adopt-%s"
  type        = "CNAME"
  value       = "example.com"
  accelerated = true
}

resource "bunnynet_pullzone" "adopted" {
  existing_id = bunnynet_dns_record.accelerated.accelerated_pullzone

  origin {
    type = "DnsAccelerate"
  }

  routing {
    tier = "Standard"
  }

  cache_enabled = true
}
`

func TestAccPullzoneExistingIdResource(t *testing.T) {
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzoneExistingIdTest, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bunnynet_pullzone.adopted", "id"),
					resource.TestCheckResourceAttrSet("bunnynet_pullzone.adopted", "name"),
					resource.TestCheckResourceAttrSet("bunnynet_pullzone.adopted", "existing_id"),
					resource.TestCheckResourceAttr("bunnynet_pullzone.adopted", "cache_enabled", "true"),
				),
			},
		},
	})
}

const configPullzoneNoNameNoExistingId = `
resource "bunnynet_pullzone" "test" {
  origin {
    type = "OriginUrl"
    url = "https://bunny.net"
  }

  routing {
    tier = "Standard"
  }
}
`

func TestAccPullzoneNoNameNoExistingId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      configPullzoneNoNameNoExistingId,
				ExpectError: regexp.MustCompile(`Either "name" or "existing_id" must be configured`),
			},
		},
	})
}
