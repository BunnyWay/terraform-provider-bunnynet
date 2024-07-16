resource "bunny_pullzone_hostname" "bunny" {
  pullzone  = bunny_pullzone.example.id
  name      = "my-pullzone.b-cdn.net"
  force_ssl = false
}

resource "bunny_pullzone_hostname" "custom" {
  pullzone  = bunny_pullzone.example.id
  name      = "cdn.example.com"
  force_ssl = true
}
