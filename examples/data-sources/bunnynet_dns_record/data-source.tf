data "bunnynet_dns_record" "A" {
  zone = data.bunnynet_dns_zone.example.id
  type = "A"
  name = "www"

  # id is optional, can be used to distinguish between records with the same name and type
  id = 123456
}

output "record" {
  value = data.bunnynet_dns_record.A
}
