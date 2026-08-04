// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const configPurgeUrlActionTest = `
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

action "bunnynet_purge_url" "test" {
  config {
    url = "https://test-acceptance-%s.b-cdn.net/index.html"
  }
}

resource "terraform_data" "purge" {
  depends_on = [bunnynet_pullzone.test]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bunnynet_purge_url.test]
    }
  }
}
`

func TestAccPurgeUrlAction(t *testing.T) {
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPurgeUrlActionTest, testKey, testKey),
			},
		},
	})
}
