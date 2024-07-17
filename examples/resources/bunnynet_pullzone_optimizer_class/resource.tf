resource "bunnynet_pullzone_optimizer_class" "thumbnail" {
  pullzone   = bunnynet_pullzone.example.id
  name       = "thumbnail"
  width      = 200
  height     = 300
  brightness = 15
}
