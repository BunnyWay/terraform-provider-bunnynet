resource "bunny_pullzone" "test" {
  name = "test"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }
}

resource "bunny_pullzone_optimizer_class" "thumbnail" {
  pullzone = bunny_pullzone.test.id
  name     = "thumbnail"

  width      = 200
  height     = 300
  brightness = 15
}
