# docker.io public images
data "bunnynet_compute_container_imageregistry" "dockerhub" {
  registry = "DockerHub"
  username = ""
}

# ghcr.io public images
data "bunnynet_compute_container_imageregistry" "github" {
  registry = "GitHub"
  username = ""
}

# ghcr.io private images
data "bunnynet_compute_container_imageregistry" "my-org" {
  registry = "GitHub"
  username = "my-org"
}
