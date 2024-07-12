resource "bunny_stream_collection" "test" {
  library = bunny_stream_library.test.id
  name    = "test collection"
}
