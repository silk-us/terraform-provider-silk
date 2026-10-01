---
page_title: "Silk Provider"
description: |-
  The Silk provider exposes resources to interact with a Silk SDP server.
---

# Silk Provider

The Silk provider exposes resources to interact with a Silk SDP server.

## Example Usage

```hcl
terraform {
  required_providers {
    silk = {
      source  = "silk-us/silk"
      version = "~> 1.2"
    }
  }
}

provider "silk" {}

resource "silk_volume_group" "Silk-Volume-Group" {
  name                 = "TerraformVolumeGroup"
  quota_in_gb          = 30
  enable_deduplication = true
  description          = "Created through Terraform"
}
```

## Notes

* Object names on the Silk server are limited to 32 characters.
* Resources track their object by its SDP ID (`obj_id`). If an object is renamed outside of Terraform, the next plan renames it back instead of treating it as deleted. For resources where the name forces a replacement, such as Capacity Policies, the next plan replaces the object instead.

## Authentication

The Silk provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

* Static credentials
* Environment variables

### Static credentials

Static credentials can be provided by adding a `server`, `username` and
`password` in-line in the Silk provider block:

```hcl
provider "silk" {
  server   = "192.0.1.10"
  username = "admin"
  password = "admin"
}
```

### Environment variables

You can provide your credentials via the `SILK_SDP_SERVER`,
`SILK_SDP_USERNAME` and `SILK_SDP_PASSWORD` environment variables,
representing your Silk server IP address, username and password,
respectively.

```sh
export SILK_SDP_SERVER="192.0.1.10"
export SILK_SDP_USERNAME="admin"
export SILK_SDP_PASSWORD="admin"
```

```hcl
provider "silk" {}
```

## Schema

### Optional

* `server` - (Optional) The IP Address of a Silk server. May also be sourced from the `SILK_SDP_SERVER` environment variable.
* `username` - (Optional) The username used to authenticate against the Silk server. May also be sourced from the `SILK_SDP_USERNAME` environment variable.
* `password` - (Optional) The password used to authenticate against the Silk server. May also be sourced from the `SILK_SDP_PASSWORD` environment variable.
