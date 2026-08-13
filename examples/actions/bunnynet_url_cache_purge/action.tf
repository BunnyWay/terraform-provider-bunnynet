# purge all URLs under /assets/
action "bunnynet_url_cache_purge" "website-assets" {
  config {
    url = "https://www.example.com/assets/"
  }
}

# purge only /contact/, but not its sub-resources
action "bunnynet_url_cache_purge" "website-assets" {
  config {
    url        = "https://www.example.com/contact/"
    exact_path = true
  }
}

# purge all URLs prefixed with `/test-`
action "bunnynet_url_cache_purge" "website-assets" {
  config {
    url = "https://www.example.com/test-*"
  }
}
