resource "random_password" "subuser" {
  length  = 17
  upper   = true
  lower   = true
  numeric = true
  special = true
}

resource "bunnynet_account_subuser" "test" {
  firstname   = "John"
  lastname    = "Doe"
  email       = "john@example.com"
  password    = random_password.subuser.result
  permissions = ["zones", "billing"]
}
