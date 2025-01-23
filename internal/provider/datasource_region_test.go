// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegionDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.bunnynet_region.LJ", "region_code", "LJ"),
					resource.TestCheckResourceAttr("data.bunnynet_region.LJ", "name", "EU: Ljubljana, SI"),
				),
			},
		},
	})
}

const testAccRegionDataSourceConfig = `
data "bunnynet_region" "LJ" {
  region_code = "LJ"
}
`
