resource "bunny_video_library" "test" {
  name = "test library"
}

resource "bunny_video_collection" "test" {
  library = bunny_video_library.test.id
  name    = "test collection"
}
