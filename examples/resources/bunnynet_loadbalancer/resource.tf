resource "bunnynet_loadbalancer" "test" {
  name            = "test"
  method          = "Latency"
  sticky_sessions = "Cookie"
}
