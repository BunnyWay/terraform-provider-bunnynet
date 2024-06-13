resource "bunny_pullzone" "test" {
  name = "test"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }
}

resource "bunny_pullzone_hostname" "bunny" {
  pullzone  = bunny_pullzone.test.id
  name      = "test.b-cdn.net"
  force_ssl = false
}

resource "bunny_pullzone_hostname" "custom" {
  pullzone  = bunny_pullzone.test.id
  name      = "cdn.example.com"
  force_ssl = true
}
