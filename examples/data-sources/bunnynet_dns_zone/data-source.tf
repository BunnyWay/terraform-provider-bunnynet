data "bunnynet_dns_zone" "example" {
  domain = "example.org"
}

output "nameservers" {
  value = [
    data.bunnynet_dns_zone.example.nameserver1,
    data.bunnynet_dns_zone.example.nameserver2,
  ]
}
