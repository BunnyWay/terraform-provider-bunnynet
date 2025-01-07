// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configDnsRecordPZTest = `
resource "bunnynet_pullzone" "pullzone" {
  name = "test-acceptance-%s"

  origin {
    type = "OriginUrl"
    url  = "https://bunny.net"
  }

  routing {
    tier = "Standard"
  }
}

data "bunnynet_dns_zone" "domain" {
  domain = "terraform.internal"
}

resource "bunnynet_dns_record" "record" {
  zone      = data.bunnynet_dns_zone.domain.id
  name      = "test22"
  type      = "PullZone"
  value     = bunnynet_pullzone.pullzone.name
  link_name = bunnynet_pullzone.pullzone.id
}
`

func TestAccDnsRecordResourcePZ(t *testing.T) {
	testKey := generateRandomString(12)
	config := fmt.Sprintf(configDnsRecordPZTest, testKey)
	pzName := fmt.Sprintf("test-acceptance-%s", testKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bunnynet_pullzone.pullzone", "name", pzName),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "type", "PullZone"),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "value", pzName),
				),
			},
		},
	})
}
