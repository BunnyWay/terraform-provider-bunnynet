resource "bunny_dns_record" "MX" {
  zone = bunny_dns_zone.example.id

  name     = ""
  type     = "MX"
  value    = "mail.example.com."
  priority = 1
}
