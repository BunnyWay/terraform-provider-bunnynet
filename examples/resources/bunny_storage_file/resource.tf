resource "bunny_storage_file" "homepage" {
  zone    = bunny_storage_zone.test.id
  path    = "index.html"
  content = "<h1>Hello world</h1>"
}
