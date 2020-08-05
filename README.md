# Silk Provider for Terraform

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">


# Installation

Requirements: Terraform has been successfully [installed](https://learn.hashicorp.com/terraform/getting-started/install.html).

# Documentation

# Example 

```hcl
provider "silk" {
  server = "192.0.1.60"
  username = "username"
  password = "password"
}

resource "silk_volume" "Silk-Volume" {
  name = "Terraform2"
  size_in_gb = 30
  volume_group_name = "TerraformVolumeGroup"
  vmware = true
  description = "Created through TF"
  read_only = false
  host_mapping = ["ExampleHostName]
  host_group_mapping = ["ExampleHostGroupName]
  allow_destroy = true
}

resource "silk_volume_group" "Silk-Volume-Group" {
  name = "TerraformVolumeGroup2"
  quota_in_gb = 30
  enable_deduplication = true
  description = "Crated through TF"
}

resource "silk_host" "Silk-Host" {
  name = "TerraformHost"
  host_type = "Linux"
  pwwn = ["20:36:44:78:66:77:ab:10", "30:36:44:78:66:77:ab:10", "50:36:44:78:66:77:ab:10"]
}

resource "silk_host_group" "Silk-Host-Group" {
  name = "TerraformHostGroup"
  description = "Updated through Terraform"
  allow_different_host_types = true
  host_mapping = ["ExampleHostName", "ExampleHostName"]

}
```