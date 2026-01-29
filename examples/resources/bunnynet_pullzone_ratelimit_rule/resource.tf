resource "bunnynet_pullzone_ratelimit_rule" "wplogin" {
  pullzone        = bunnynet_pullzone.test.id
  name            = "WordPress Login"
  description     = "WordPress Login"
  transformations = ["LOWERCASE", "NORMALIZEPATH", "URLDECODE"]

  condition {
    variable = "REQUEST_URI"
    operator = "BEGINSWITH"
    value    = "/wp-login.php"
  }

  limit {
    requests = 2
    interval = 10 # seconds
  }

  response {
    interval = 3600 # seconds
  }
}
