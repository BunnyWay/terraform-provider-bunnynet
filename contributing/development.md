# Development Environment Setup

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Quick Start

To contribute tu BonnyNet Terraform Provider, you'll first need [Go](http://www.golang.org)
installed on your machine (version 1.18+ is _required_). In addition you need to setup a [GOPATH](http://golang.org/doc/code.html#GOPATH).
Finally, add `$GOPATH/bin` to your `$PATH`.

## Using the BunntNet Terraform Provider

With Terraform v0.14 and later, [development overrides for provider developers](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers) can be leveraged in order to use the provider built from source.

Populate a Terraform CLI configuration file (`~/.terraformrc` for
all platforms other than Windows; `terraform.rc` in the `%APPDATA%` directory
when using Windows) with at least the following options:

```
provider_installation {
  dev_overrides {
    "bunnynet/bunnynet" = "<GOPATH>/src/github.com/BunnyWay/terraform-provider-bunnynet"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

You will need to replace `<GOPATH>` with the **full path** to your GOPATH where
the repository lives, no `~` shorthand.

Once you have this file in place, you can run `make build-dev` which will
build a development version of the binary in the repository that Terraform
will use instead of the version from the remote registry.
