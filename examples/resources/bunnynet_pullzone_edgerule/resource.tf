resource "bunnynet_pullzone_edgerule" "redirect_admin" {
  enabled     = true
  pullzone    = bunnynet_pullzone.example.id
  description = "Redirect to homepage"

  actions = [
    {
      type       = "Redirect"
      parameter1 = "https://www.example.com"
      parameter2 = "302"
      parameter3 = null
    }
  ]

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
