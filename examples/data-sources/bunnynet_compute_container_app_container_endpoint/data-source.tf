data "bunnynet_compute_container_app_container_endpoint" "nginx_http" {
  app       = vars.app_id
  container = data.bunnynet_compute_container_app_container_endpoint.nginx.id
  name      = "http"
}

output "url" {
  value = "https://${data.bunnynet_compute_container_app_container_endpoint.nginx_http.public_host}"
}
