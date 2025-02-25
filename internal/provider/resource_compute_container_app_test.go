// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"
)

const configComputeContainerAppTest = `
data "bunnynet_compute_container_imageregistry" "dockerhub" {
  registry = "DockerHub"
  username = ""
}

resource "bunnynet_compute_container_app" "test" {
	name = "test-acceptance-%s"
	regions_allowed = ["DE"]
	regions_required = ["DE"]

	container {
		name = "echo"
		image_registry = data.bunnynet_compute_container_imageregistry.dockerhub.id
		image_namespace = "hashicorp"
		image_name = "http-echo"
		image_tag = "latest"

		endpoint {
			name = "cdn"
			type = "CDN"

			cdn {
				origin_ssl = false
			}

			port {
				container = 5678
			}
		}
	}
}
`

func TestAccComputeContainerAppResource(t *testing.T) {
	resourceName := "bunnynet_compute_container_app.test"
	testKey := generateRandomString(12)
	config := fmt.Sprintf(configComputeContainerAppTest, testKey)

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
				ImportStateIdFunc: testAccComputeContainerAppImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccComputeContainerAppImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}
