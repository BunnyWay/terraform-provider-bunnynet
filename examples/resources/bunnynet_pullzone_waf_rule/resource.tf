resource "bunnynet_pullzone_waf_rule" "test" {
  pullzone    = bunnynet_pullzone.test.id
  name        = "WordPress Login"
  description = "Challenge any request"

  condition {
    variable        = "REQUEST_URI"
    operator        = "BEGINSWITH"
    value           = "/wp-login.php"
    transformations = ["LOWERCASE", "NORMALIZEPATH", "URLDECODE"]
  }

  response {
    action = "Challenge"
  }
}
