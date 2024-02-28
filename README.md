# Silk Terraform Provider

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

# Quick Start

Installation and Usage information can be found in the Silk Terraform Provider [Quick Start Guide](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/quick_start.md).

# Release

Place the appropriate binary from the release into the localdomain location for your terraform plugins. For example:
```
mv terraform-provider-silk_1.0.9_linux_amd64 ~/.terraform.d/plugins/localdomain/provider/silk/1.0.9/linux_amd64/terraform-provider-silk
```

# Build

Makefile is included, simply unzip the file and run `make` or `make install`. Requires `Go`.


# Documentation

* [Provider](https://github.com/silk-us/silk-terraform-provider/tree/master/docs)
* [silk_host](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_host.md)
* [silk_host_group](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_host_group.md)
* [silk_volume](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_volume.md)
* [silk_volume_group](https://github.com/silk-us/silk-terraform-provider/blob/master/docs/silk_volume_group.md)
* [silk_retention_policy](https://github.com/silk-us/terraform-provider-silk/blob/master/docs/silk_retention_policy.md)
* [silk_capacity_policy](https://github.com/silk-us/terraform-provider-silk/blob/master/docs/silk_capacity_policy.md)
