# Silk Terraform Provider

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">


# Quick Start

Installation and Usage information can be found in the Silk Terraform Provider [Quick Start Guide](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/quick_start.md).

# Build

Makefile is included, simply unzip the file and run `make` or `make install`. Requires `Go`.

# Example

```hcl
provider "silk" {}

resource "silk_volume_group" "Silk-Volume-Group" {
  name = "TerraformVolumeGroup"
  quota_in_gb = 30
  enable_deduplication = true
  description = "Crated through TF"
}
```

For Terraform version 0.13 or later, you will want to move the binary to an appropriate path and to add a `required_providers` statement for this provider. For example (for version 1.0.9):

```
mv terraform-provider-silk ~/.terraform.d/plugins/localdomain/provider/silk/1.0.9/linux_amd64
```

And then add the provider statement:

```hcl
terraform {
  required_providers {
    silk = {
      source  = "localdomain/provider/silk"
      version = "1.0.9"
    }
  }
}
```

# Documentation

* [Provider](https://github.com/silk-us/silk-terraform-provider/tree/master/docs)
* [silk_host](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_host.md)
* [silk_host_group](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_host_group.md)
* [silk_volume](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_volume.md)
* [silk_volume_group](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_volume_group.md)
* [silk_retention_policy](https://github.com/silk-us/terraform-provider-silk/blob/master/docs/silk_retention_policy.md)
* [silk_capacity_policy](https://github.com/silk-us/terraform-provider-silk/blob/master/docs/silk_capacity_policy.md)
