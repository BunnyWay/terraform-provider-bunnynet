resource "bunny_stream_video" "example" {
  library    = bunny_stream_library.example.id
  collection = bunny_stream_collection.example.id

  title       = "Big Buck Bunny (2008)"
  description = "Big Buck Bunny (code-named Project Peach) is a 2008 animated comedy short film featuring animals of the forest, made by the Blender Institute, part of the Blender Foundation."

  chapters = [
    {
      title = "Title"
      start = "00:00"
      end   = "00:05"
    },
    {
      title = "Film"
      start = "00:05"
      end   = "08:15"
    },
    {
      title = "Credits"
      start = "08:15"
      end   = "10:19"
    },
    {
      title = "Outro"
      start = "10:19"
      end   = "10:35"
    },
  ]

  moments = [
    {
      label     = "Bird gets hit with an acorn"
      timestamp = "00:25"
    },
    {
      label     = "Buck gets hit with an acorn"
      timestamp = "02:43"
    },
  ]
}
