resource "bunnynet_dns_record" "MX" {
  zone = bunnynet_dns_zone.example.id

  name  = ""
  type  = "A"
  value = "192.0.2.33"
}
