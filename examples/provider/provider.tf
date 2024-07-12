terraform {
  required_providers {
    bunny = {
      source = "terraform-provider-bunny.b-cdn.net/bunny/bunny"
    }
  }
}

provider "bunny" {
  api_key = "00000000-0000-0000-0000-000000000000"
}

resource "bunny_pullzone" "example" {
  name = "example"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }

  routing {
    tier = "Standard"
  }
}
