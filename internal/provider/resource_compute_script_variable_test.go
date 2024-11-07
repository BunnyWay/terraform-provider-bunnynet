// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configComputeScriptVariableTest = `
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

resource "bunnynet_compute_script_variable" "test" {
  script        = bunnynet_compute_script.test.id
  name          = "VAR_%s"
  default_value = "%s"
  required      = true
}
`

func TestAccComputeScriptVariableResource(t *testing.T) {
	scriptResource := "bunnynet_compute_script.test"
	variableResource := "bunnynet_compute_script_variable.test"

	scriptName := generateRandomString(12)
	variableName := generateRandomString(12)
	variableValue := generateRandomString(12)
	config := fmt.Sprintf(configComputeScriptVariableTest, scriptName, variableName, variableValue)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(scriptResource, "name", fmt.Sprintf("test-acceptance-%s", scriptName)),
					resource.TestCheckResourceAttr(variableResource, "name", fmt.Sprintf("VAR_%s", variableName)),
					resource.TestCheckResourceAttr(variableResource, "default_value", variableValue),
				),
			},
			{
				ResourceName:      variableResource,
				ImportState:       true,
				ImportStateIdFunc: testAccComputeScriptVariableImportStateIdFunc(variableResource),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccComputeScriptVariableImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s|%s", rs.Primary.Attributes["script"], rs.Primary.Attributes["name"]), nil
	}
}
