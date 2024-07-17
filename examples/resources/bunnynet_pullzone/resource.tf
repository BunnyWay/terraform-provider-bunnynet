resource "bunnynet_pullzone" "example" {
  name = "my-website"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }

  routing {
    tier = "Standard"
  }
}
