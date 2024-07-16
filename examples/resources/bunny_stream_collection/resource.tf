resource "bunny_stream_collection" "example" {
  library = bunny_stream_library.example.id
  name    = "example collection"
}
