resource "bunnynet_pullzone_waf_rule" "test" {
  pullzone        = bunnynet_pullzone.test.id
  name            = "WordPress Login"
  description     = "Challenge any request"
  transformations = ["LOWERCASE", "NORMALIZEPATH", "URLDECODE"]

  condition {
    variable = "REQUEST_URI"
    operator = "BEGINSWITH"
    value    = "/wp-login.php"
  }

  response {
    action = "Challenge"
  }
}
