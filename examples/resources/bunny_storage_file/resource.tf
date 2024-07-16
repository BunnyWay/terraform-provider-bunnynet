resource "bunny_storage_file" "homepage" {
  zone = bunny_storage_zone.example.id
  path = "index.html"

  ## file contents
  # either
  content = "<h1>Hello world</h1>"
  # or
  source = "data/index.html"
}
