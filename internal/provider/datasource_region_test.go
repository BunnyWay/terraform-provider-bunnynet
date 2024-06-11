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
					resource.TestCheckResourceAttr("data.bunny_region.LJ", "id", "98"),
					resource.TestCheckResourceAttr("data.bunny_region.LJ", "name", "EU: Ljubljana, SI"),
				),
			},
		},
	})
}

const testAccRegionDataSourceConfig = `
data "bunny_region" "LJ" {
  region_code = "LJ"
}
`
