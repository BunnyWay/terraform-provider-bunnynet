resource "bunnynet_pullzone_hostname" "bunnynet" {
  pullzone  = bunnynet_pullzone.example.id
  name      = "my-pullzone.b-cdn.net"
  force_ssl = false
}

resource "bunnynet_pullzone_hostname" "custom" {
  pullzone  = bunnynet_pullzone.example.id
  name      = "cdn.example.com"
  force_ssl = true
}
