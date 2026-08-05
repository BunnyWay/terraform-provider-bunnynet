# Purge the entire cache after an edge rule changes,
# so already-cached responses pick up the new behavior.
action "bunnynet_pullzone_purge_cache" "example" {
  config {
    pullzone = bunnynet_pullzone.example.id
  }
}

resource "bunnynet_pullzone_edgerule" "security_headers" {
  enabled     = true
  pullzone    = bunnynet_pullzone.example.id
  description = "Add security headers"

  actions = [
    {
      type       = "SetResponseHeader"
      parameter1 = "X-Frame-Options"
      parameter2 = "DENY"
      parameter3 = null
    }
  ]

  match_type = "MatchAny"
  triggers = [
    {
      type       = "Url"
      match_type = "MatchAny"
      patterns   = ["*"]
      parameter1 = null
      parameter2 = null
    }
  ]

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.bunnynet_pullzone_purge_cache.example]
    }
  }
}

# Purge only cached responses tagged "assets" via the CDN-Tag
# response header. CDN-* headers cannot be set by edge rules, so
# the origin itself must send the header.
#
# Besides lifecycle action_trigger, actions can also be invoked
# on demand from the CLI:
#   terraform apply -invoke='action.bunnynet_pullzone_purge_cache.assets'
action "bunnynet_pullzone_purge_cache" "assets" {
  config {
    pullzone  = bunnynet_pullzone.example.id
    cache_tag = "assets"
  }
}
