provider "silk" {
  server = "10.30.1.60"
  username = "admin"
  password = "admin"
}

resource "silk_volume" "Silk-Volume" {
  name = "Terraform2"
  size_in_gb = 30
  volume_group_name = "TerraformVolumeGroup"
  vmware = true
  description = "Created through TF"
  read_only = false
  allow_destroy = true
}

resource "silk_volume_group" "Create-Silk-Volume" {
  name = "TerraformVolumeGroup2"
  quota_in_gb = 30
  enable_deduplication = true
  description = "Crated through TF"
}

resource "silk_host" "Silk-Host" {
  name = "TerraformHost"
  host_type = "Linux"
  pwwn = ["20:36:44:78:66:77:ab:10", "30:36:44:78:66:77:ab:10", "50:36:44:78:66:77:ab:10", "10:96:44:78:66:77:ab:10"]
}

resource "silk_host_group" "Silk-Host-Group" {
  name = "TerraformHostGroup"
  description = "Updated through Terraform"
  allow_different_host_types = true

}