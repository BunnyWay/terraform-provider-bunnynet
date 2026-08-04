action "bunnynet_purge_url" "homepage" {
  config {
    url = "https://myzone.b-cdn.net/index.html"
  }
}

resource "bunnynet_storage_file" "homepage" {
  zone    = bunnynet_storage_zone.example.id
  path    = "index.html"
  content = "<h1>Hello world</h1>"

  lifecycle {
    action_trigger {
      events  = [after_update]
      actions = [action.bunnynet_purge_url.homepage]
    }
  }
}
