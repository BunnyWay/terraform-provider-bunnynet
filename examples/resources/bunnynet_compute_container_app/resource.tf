resource "bunnynet_compute_container_app" "app" {
  name    = "my-app"
  version = 2

  autoscaling_min = 1
  autoscaling_max = 3

  regions_allowed  = ["DE", "SG", "NY"]
  regions_required = ["DE"]

  container {
    name            = "app"
    image_registry  = bunnynet_compute_container_imageregistry.github.id
    image_namespace = "my-org"
    image_name      = "my-app"
    image_tag       = "2.12.6"

    endpoint {
      name = "app"
      type = "CDN"

      cdn {
        origin_ssl = false

        sticky_sessions {
          headers = ["X-Forwarded-For"]
        }
      }

      port {
        container = 8080
        protocols = ["TCP", "UDP"]
      }
    }

    env {
      name  = "APP_ENV"
      value = "prod"
    }

    env {
      name  = "LISTEN_PORT"
      value = "3000"
    }
  }

  container {
    name            = "sidecar"
    image_registry  = bunnynet_compute_container_imageregistry.github.id
    image_namespace = "my-org"
    image_name      = "my-sidecar"
    image_tag       = "1.3.7"
  }
}
