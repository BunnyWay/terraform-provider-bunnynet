resource "bunnynet_pullzone_shield" "test" {
  pullzone = bunnynet_pullzone.test.id
  tier     = "Standard"

  ddos {
    level = "Medium"
    mode  = "Block"
  }

  waf {
    enabled = true
    mode    = "Block"
  }
}
