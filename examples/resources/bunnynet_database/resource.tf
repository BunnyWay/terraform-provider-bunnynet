resource "bunnynet_database" "test" {
  name            = "mydb"
  regions_primary = ["DE", "NY", "SG"]
  regions_replica = ["LA", "UK", "SYD"]
}
