resource "bunnynet_pullzone_hostname" "bunnynet" {
  pullzone  = bunnynet_pullzone.example.id
  name      = "my-pullzone.b-cdn.net"
  force_ssl = false
}

resource "bunnynet_pullzone_hostname" "custom" {
  pullzone    = bunnynet_pullzone.example.id
  name        = "cdn.example.com"
  tls_enabled = true
  force_ssl   = true

  # Uploads a custom certificate. To use a managed certificate (Let's Encrypt), omit both attributes.
  certificate     = file("cdn.example.com.cert")
  certificate_key = file("cdn.example.com.key")
}
