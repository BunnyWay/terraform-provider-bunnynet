resource "bunny_dns_zone" "example" {
  domain = "example.com"

  nameserver_custom = true
  nameserver1       = "ns1.example.com"
  nameserver2       = "ns2.example.com"
  soa_email         = "hostmaster@example.com"

  log_enabled          = true
  log_anonymized       = true
  log_anonymized_style = "OneDigit"
}
