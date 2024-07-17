resource "bunnynet_stream_collection" "example" {
  library = bunnynet_stream_library.example.id
  name    = "example collection"
}
