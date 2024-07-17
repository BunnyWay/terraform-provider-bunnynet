resource "bunnynet_pullzone_edgerule" "block_admin" {
  enabled     = true
  pullzone    = bunnynet_pullzone.example.id
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
