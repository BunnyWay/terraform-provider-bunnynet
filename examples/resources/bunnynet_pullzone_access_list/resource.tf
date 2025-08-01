resource "bunnynet_pullzone_access_list" "hetzner" {
  pullzone = bunnynet_pullzone.test.id
  name     = "VIP customers"
  action   = "Allow"
  type     = "CIDR"
  entries = [
    "192.0.2.0/24",
    "2001:db8:cafe::/48"
  ]
}
