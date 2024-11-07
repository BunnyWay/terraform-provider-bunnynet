resource "bunnynet_compute_script_variable" "APP_ENV" {
  script        = bunnynet_compute_script.test.id
  name          = "APP_ENV"
  default_value = "prod"
  required      = true
}
