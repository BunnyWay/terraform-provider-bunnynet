// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configComputeScriptSecretTest = `
resource "bunnynet_compute_script" "test" {
  type    = "standalone"
  name    = "test-acceptance-%s"
  content = <<-CODE
	import * as BunnySDK from "https://esm.sh/@bunny.net/edgescript-sdk@0.10.0";
    import process from "node:process"

    BunnySDK.net.http.serve(async (request: Request): Response | Promise<Response> => {
        return new Response('<p>APP_SECRET: ' + process.env.SECRET_%s + '</p>', {headers: {"Content-Type": "text/html"}});
    });
  CODE
}

resource "bunnynet_compute_script_secret" "test" {
  script = bunnynet_compute_script.test.id
  name   = "SECRET_%s"
  value  = "%s"
}
`

func TestAccComputeScriptSecretResource(t *testing.T) {
	scriptResource := "bunnynet_compute_script.test"
	secretResource := "bunnynet_compute_script_secret.test"

	scriptName := generateRandomString(12)
	secretName := generateRandomString(4)
	secretValue := generateRandomString(12)
	config := fmt.Sprintf(configComputeScriptSecretTest, scriptName, secretName, secretName, secretValue)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(scriptResource, "name", fmt.Sprintf("test-acceptance-%s", scriptName)),
					resource.TestCheckResourceAttr(secretResource, "name", fmt.Sprintf("SECRET_%s", secretName)),
					resource.TestCheckResourceAttr(secretResource, "value", secretValue),
				),
			},
			{
				ResourceName:            secretResource,
				ImportState:             true,
				ImportStateIdFunc:       testAccComputeScriptSecretImportStateIdFunc(secretResource),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func testAccComputeScriptSecretImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["script"], rs.Primary.Attributes["name"]), nil
	}
}
