resource "bunnynet_loadbalancer_origin" "DK" {
  loadbalancer = bunnynet_loadbalancer.test.id
  group        = bunnynet_loadbalancer_group.EU.id
  name         = "DK"
  weight       = 1
  type         = "OriginUrl"
  url          = "https://origin-dk.example.net"
}
