resource "bunnynet_loadbalancer_origin" "DK" {
  loadbalancer = bunnynet_loadbalancer.test.id
  group        = bunnynet_loadbalancer_group.EU.id
  group_weight = 1
  name         = "DK"
  type         = "OriginUrl"
  url          = "https://origin-dk.example.net"
}
