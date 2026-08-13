action "bunnynet_pullzone_cache_purge" "website-assets" {
  config {
    pullzone = bunnynet_pullzone.website.id
    tag      = "assets"
  }
}
