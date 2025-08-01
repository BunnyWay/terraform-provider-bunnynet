resource "bunnynet_pullzone" "test" {
  // ...
}

data "bunnynet_pullzone_access_lists" "curated" {
  pullzone = bunnynet_pullzone.test.id
  custom   = false
}

resource "bunnynet_pullzone_shield" "test" {
  // ...
  pullzone = bunnynet_pullzone.test.id

  access_list {
    id     = data.bunnynet_pullzone_access_lists.curated.data["VPN Providers"].id
    action = "Block"
  }

  access_list {
    id     = data.bunnynet_pullzone_access_lists.curated.data["TOR Exit Nodes"].id
    action = "Block"
  }
}
