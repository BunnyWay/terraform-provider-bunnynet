resource "bunnynet_compute_script_secret" "APP_SECRET" {
  script = bunnynet_compute_script.test.id
  name   = "APP_ENV"
  value  = ""
}
