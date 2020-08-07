# Silk Provider

The Silk Provider exposes resources to interact with a Silk server.

## Example Usage

``` hcl
provider "silk" {}

resource "silk_volume_group" "Silk-Volume-Group" {
  name = "TerraformVolumeGroup"
  quota_in_gb = 30
  enable_deduplication = true
  description = "Crated through TF"
}
```

## Authentication

The Silk provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

* Static credentials
* Environment variables

### Static credentials 

Static credentials can be provided by adding an `server` , `username` and `password` in-line in the Silk provider block:

Usage:

``` hcl
provider "silk" {
  server     = "192.0.1.601"
  username    = "admin"
  password    = "admin"
}
```

### Environment variables

You can provide your credentials via the `SILK_SDP_SERVER` , `SILK_SDP_USERNAME` and
`SILK_SDP_PASSWORD` , environment variables, representing your Silk Server IP address, username
and password, respectively.

``` sh
$ export SILK_SDP_SERVER="192.0.1.011"
$ export SILK_SDP_USERNAME="admin"
$ export SILK_SDP_PASSWORD="admin"
```

``` hcl
provider "silk" {}
```

## Argument Reference

The following arguments are supported in the Silk `provider` block:

* `server` - (Optional) The IP Address of a Silk server. The value may also be sourced from the `SILK_SDP_SERVER` environment variable.

* `username` - (Optional) The username used to authenticate against the Silk server. The value may also be sourced from the `SILK_SDP_USERNAME` environment variable.

* `password` - (Optional) The password used to authenticate against the Silk Sever. The value may also be sourced from the `SILK_SDP_PASSWORD` environment variable.
