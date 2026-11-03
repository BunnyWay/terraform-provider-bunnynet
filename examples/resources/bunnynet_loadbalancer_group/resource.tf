resource "bunnynet_loadbalancer_group" "EU" {
  loadbalancer        = bunnynet_loadbalancer.test.id
  name                = "EU"
  method              = "Latency"
  priority            = 1
  healthcheck_enabled = true
  healthcheck_path    = "/health"
}
