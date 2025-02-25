data "bunnynet_compute_container_app_container" "nginx" {
  app  = vars.app_id
  name = "nginx"
}
