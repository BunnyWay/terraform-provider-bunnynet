resource "bunny_pullzone_optimizer_class" "thumbnail" {
  pullzone   = bunny_pullzone.example.id
  name       = "thumbnail"
  width      = 200
  height     = 300
  brightness = 15
}
