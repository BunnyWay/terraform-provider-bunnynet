// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configComputeScriptTest = `
resource "bunnynet_compute_script" "test" {
  type    = "standalone"
  name    = "test-acceptance-%s"
  content = <<-CODE
    import * as BunnySDK from "https://esm.sh/@bunny.net/edgescript-sdk@0.10.0";

    BunnySDK.net.http.serve(async (request: Request): Response | Promise<Response> => {
        return new Response('<h1>Hello world!</h1>', {headers: {"Content-Type": "text/html"}});
    });
  CODE
}
`

func TestAccComputeScriptResource(t *testing.T) {
	resourceName := "bunnynet_compute_script.test"
	testKey := generateRandomString(12)
	config := fmt.Sprintf(configComputeScriptTest, testKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acceptance-%s", testKey)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccComputeScriptImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccComputeScriptImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}
