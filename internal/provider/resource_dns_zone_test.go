// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configDnsZoneDnssec = `
resource "bunnynet_dns_zone" "domain" {
  domain = "terraform-acc-%s.internal"
  dnssec_enabled = %t
}
`

func TestAccDnsZoneDnssecDisabled(t *testing.T) {
	testKeyOk := generateRandomString(4)
	configOk := fmt.Sprintf(configDnsZoneDnssec, testKeyOk, false)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configOk,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("domain"), knownvalue.StringExact(fmt.Sprintf("terraform-acc-%s.internal", testKeyOk))),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_algorithm"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_public_key"), knownvalue.StringFunc(func(v string) error {
						if v != "" {
							return errors.New("value not empty")
						}
						return nil
					})),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_digest"), knownvalue.StringFunc(func(v string) error {
						if v != "" {
							return errors.New("value not empty")
						}
						return nil
					})),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_digest_type"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_flags"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_keytag"), knownvalue.Int64Exact(0)),
				},
			},
		},
	})
}

func TestAccDnsZoneDnssecEnabled(t *testing.T) {
	testKeyOk := generateRandomString(4)
	configOk := fmt.Sprintf(configDnsZoneDnssec, testKeyOk, true)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configOk,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("domain"), knownvalue.StringExact(fmt.Sprintf("terraform-acc-%s.internal", testKeyOk))),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_algorithm"), knownvalue.Int64Exact(13)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_digest"), knownvalue.StringFunc(func(v string) error {
						if len(v) != 64 {
							return errors.New("value is not a digest")
						}
						return nil
					})),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_digest_type"), knownvalue.Int64Exact(2)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_flags"), knownvalue.Int64Exact(257)),
					statecheck.ExpectKnownValue("bunnynet_dns_zone.domain", tfjsonpath.New("dnssec_keytag"), knownvalue.Int64Func(func(v int64) error {
						if v > 0 && v < 65536 {
							return nil
						}
						return errors.New("value is invalid")
					})),
				},
			},
		},
	})
}
