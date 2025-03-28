---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bunnynet_compute_script_variable Resource - terraform-provider-bunnynet"
subcategory: ""
description: |-
  This resource manages an Environment variable for a Compute Script in bunny.net.
---

# bunnynet_compute_script_variable (Resource)

This resource manages an Environment variable for a Compute Script in bunny.net.

## Example Usage

```terraform
resource "bunnynet_compute_script_variable" "APP_ENV" {
  script        = bunnynet_compute_script.test.id
  name          = "APP_ENV"
  default_value = "prod"
  required      = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `default_value` (String) The default value of the environment variable.
- `name` (String) The name of the environment variable.
- `required` (Boolean) Indicates whether the environment variable is required.
- `script` (Number) The ID of the associated compute script.

### Read-Only

- `id` (Number) The ID of the environment variable.

## Import

Import is supported using the following syntax:

```shell
terraform import bunnynet_compute_script_variable.test "1234|APP_ENV"
```
