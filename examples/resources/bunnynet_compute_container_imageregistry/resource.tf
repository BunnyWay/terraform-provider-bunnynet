resource "bunnynet_compute_container_imageregistry" "github" {
  registry = "GitHub"
  username = "my-github-username"
  token    = "ghp_abc1234d"
}
