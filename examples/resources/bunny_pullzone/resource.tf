resource "bunny_pullzone" "test" {
  name = "test"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }
}
