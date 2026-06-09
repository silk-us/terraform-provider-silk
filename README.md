# Silk Terraform Provider

The Silk Terraform Provider exposes resources to manage a Silk SDP server —
volumes, volume groups, hosts, host groups, retention/capacity policies, and
thin clones.

- Terraform Registry: https://registry.terraform.io/providers/silk-us/silk/latest
- Built on the [silk-sdp-go-sdk](https://github.com/silk-us/silk-sdp-go-sdk)

## Installation

The provider is published to the [Terraform Registry](https://registry.terraform.io/providers/silk-us/silk/latest).
Add it to your configuration and run `terraform init` — Terraform downloads it
automatically:

```hcl
terraform {
  required_providers {
    silk = {
      source  = "silk-us/silk"
      version = "~> 1.2"
    }
  }
}

provider "silk" {
  # credentials can be set here or via environment variables (see below)
}
```

## Authentication

Credentials can be supplied inline or via environment variables.

**Inline:**

```hcl
provider "silk" {
  server   = "192.0.1.10"
  username = "admin"
  password = "admin"
}
```

**Environment variables:**

```sh
export SILK_SDP_SERVER="192.0.1.10"
export SILK_SDP_USERNAME="admin"
export SILK_SDP_PASSWORD="admin"
```

```hcl
provider "silk" {}
```

On Windows (PowerShell), use `setx` to persist them:

```powershell
setx SILK_SDP_SERVER   "192.0.1.10"
setx SILK_SDP_USERNAME "admin"
setx SILK_SDP_PASSWORD "admin"
```

## Example

```hcl
resource "silk_volume_group" "example" {
  name                 = "TerraformVolumeGroup"
  quota_in_gb          = 30
  enable_deduplication = true
  description          = "Created through Terraform"
}

resource "silk_volume" "example" {
  name              = "ExampleVolume"
  size_in_gb        = 10
  volume_group_name = silk_volume_group.example.name
  description       = "Created through Terraform"
  allow_destroy     = true
}
```

## Documentation

Full provider and resource documentation:

* [Provider overview](docs/index.md)
* [silk_volume](docs/resources/volume.md)
* [silk_volume_group](docs/resources/volume_group.md)
* [silk_host](docs/resources/host.md)
* [silk_host_group](docs/resources/host_group.md)
* [silk_retention_policy](docs/resources/retention_policy.md)
* [silk_capacity_policy](docs/resources/capacity_policy.md)
* [silk_thin_clone](docs/resources/thin_clone.md)

## Building locally

A Makefile is included for local development. Requires `Go`.

```sh
make            # build
make install    # build and install into the local Terraform plugin dir
```

To build against a local checkout of the SDK (instead of the published module),
use the helper scripts:

```sh
./build-local.sh             # macOS / Linux
```
```powershell
.\build-local.ps1            # Windows
```

### Using a locally built provider

`make install` (or `build-local.* -Install`) places the binary in Terraform's
local plugin directory under the `localdomain/provider/silk` namespace. To
consume that build instead of the registry, point the `source` at that local
namespace:

```hcl
terraform {
  required_providers {
    silk = {
      source  = "localdomain/provider/silk"
      version = "1.2.6"
    }
  }
}

provider "silk" {}
```

The plugin path the install step writes to:

```
# macOS / Linux
~/.terraform.d/plugins/localdomain/provider/silk/<version>/<os>_<arch>/

# Windows
%APPDATA%\terraform.d\plugins\localdomain\provider\silk\<version>\<os>_<arch>\
```

Run `terraform init` in the directory containing your `.tf` files and Terraform
will pick up the local build. This is for development/testing — for normal use,
install from the registry as shown above.

## Releasing

Maintainer release process (tag a `vX.Y.Z` and the registry auto-publishes) is
documented in [POST_COMMIT.md](POST_COMMIT.md). First-time registry setup is in
[PUBLISHING.md](PUBLISHING.md).
