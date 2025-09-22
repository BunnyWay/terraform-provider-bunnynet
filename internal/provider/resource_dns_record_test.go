// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
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
  zone        = data.bunnynet_dns_zone.domain.id
  name        = "test22"
  type        = "PullZone"
  value       = bunnynet_pullzone.pullzone.name
  pullzone_id = bunnynet_pullzone.pullzone.id
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

const configDnsRecordWeightTest = `
data "bunnynet_dns_zone" "domain" {
  domain = "terraform.internal"
}

resource "bunnynet_dns_record" "record" {
  zone      = data.bunnynet_dns_zone.domain.id
  name      = "test-%s"
  type      = "%s"
  value     = "%s"
  weight    = %d
}
`

const configDnsRecordWeightSRVTest = `
data "bunnynet_dns_zone" "domain" {
  domain = "terraform.internal"
}

resource "bunnynet_dns_record" "record" {
  zone      = data.bunnynet_dns_zone.domain.id
  name      = "test-%s"
  type      = "SRV"
  value     = "bunny.net"
  weight    = %d
  priority  = 1
}
`

func TestAccDnsRecordResourceWeightA(t *testing.T) {
	testKeyOk := generateRandomString(4)
	configOk := fmt.Sprintf(configDnsRecordWeightTest, testKeyOk, "A", "192.0.2.1", 29)

	testKeyError := generateRandomString(4)
	configError := fmt.Sprintf(configDnsRecordWeightTest, testKeyError, "A", "192.0.2.1", 103)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configOk,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "type", "A"),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "name", fmt.Sprintf("test-%s", testKeyOk)),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "value", "192.0.2.1"),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "weight", "29"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      configError,
				ExpectError: regexp.MustCompile(`The weight must be between 0 and 100`),
			},
		},
	})
}

func TestAccDnsRecordResourceWeightCNAME(t *testing.T) {
	testKey := generateRandomString(4)
	config := fmt.Sprintf(configDnsRecordWeightTest, testKey, "CNAME", "bunny.net", 29)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`The weight attribute is only available for SRV, A and AAAA records`),
			},
		},
	})
}

func TestAccDnsRecordResourceWeightSRV(t *testing.T) {
	testKeyOk := generateRandomString(4)
	configOk := fmt.Sprintf(configDnsRecordWeightSRVTest, testKeyOk, 28301)

	testKeyError := generateRandomString(4)
	configError := fmt.Sprintf(configDnsRecordWeightSRVTest, testKeyError, 183893)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configOk,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "type", "SRV"),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "name", fmt.Sprintf("test-%s", testKeyOk)),
					resource.TestCheckResourceAttr("bunnynet_dns_record.record", "weight", "28301"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      configError,
				ExpectError: regexp.MustCompile(`The weight must be between 0 and 65535`),
			},
		},
	})
}
