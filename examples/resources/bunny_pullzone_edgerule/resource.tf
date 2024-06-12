resource "bunny_pullzone" "test" {
  name = "test"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }
}

resource "bunny_pullzone_edgerule" "block_admin" {
  enabled     = true
  pullzone    = bunny_pullzone.test.id
  action      = "BlockRequest"
  description = "Block access to admin"

  match_type = "MatchAny"
  triggers = [
    {
      type       = "Url"
      match_type = "MatchAny"
      patterns   = ["https://cdn.example.com/wp-admin/*"]
      parameter1 = null
      parameter2 = null
    }
  ]
}
