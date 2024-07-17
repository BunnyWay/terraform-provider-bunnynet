terraform {
  required_providers {
    bunnynet = {
      source = "BunnyWay/bunnynet"
    }
  }
}

provider "bunnynet" {
  api_key = "00000000-0000-0000-0000-000000000000"
}

resource "bunnynet_pullzone" "example" {
  name = "example"

  origin {
    type = "OriginUrl"
    url  = "https://192.0.2.1"
  }

  routing {
    tier = "Standard"
  }
}
