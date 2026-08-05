// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const configPullzonePurgeCacheActionTest = `
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

action "bunnynet_pullzone_purge_cache" "test" {
  config {
    pullzone = bunnynet_pullzone.test.id
  }
}

resource "terraform_data" "purge" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bunnynet_pullzone_purge_cache.test]
    }
  }
}
`

func TestAccPullzonePurgeCacheAction(t *testing.T) {
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzonePurgeCacheActionTest, testKey),
			},
		},
	})
}

const configPullzonePurgeCacheActionCacheTagTest = `
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

action "bunnynet_pullzone_purge_cache" "test" {
  config {
    pullzone  = bunnynet_pullzone.test.id
    cache_tag = "assets"
  }
}

resource "terraform_data" "purge" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bunnynet_pullzone_purge_cache.test]
    }
  }
}
`

// TestAccPullzonePurgeCacheActionCacheTag validates that a purge scoped to a
// cache tag is accepted by the API. Evicting tagged responses requires the
// origin to send the CDN-Tag response header (CDN-* headers cannot be set by
// edge rules), which is out of scope for this test.
func TestAccPullzonePurgeCacheActionCacheTag(t *testing.T) {
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPullzonePurgeCacheActionCacheTagTest, testKey),
			},
		},
	})
}
