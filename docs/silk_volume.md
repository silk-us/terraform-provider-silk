## silk_volume

Manage a Volume on the Silk Server.

## Example Usage

``` hcl
resource "silk_volume_group" "Silk-Volume-Group" {
  name = "TerraformVolumeGroup"
  description = "Crated through TF"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Volume.w
* `size_in_gb` - (Required) The size, in GB, of the Volume.
* `volume_group_name` - (Required) The name of the Volume Group that the Volume should be added to.
* `vmware` - (Optional) This value corresponds to the 'VMware support' checkbox in the UI and specifies whether to enable VMFS. This value can be updated after the initial creation. Default is false.
* `description` - (Required) A description of the Volume
* `read_only` - (Optional) This value corresponds to the 'Exposure Type' radio button in the UI and specifies whether the volume should be 'Read/Write' or 'Read Only'. Default is false.
* `allow_destroy` - (Optional) When set to true, this value will prevent the volume from being destroyed through Terraform. Default is false.
* `host_mapping` - (Optional) A list of Hosts the Volume is mapped to.
* `host_group_mapping` - (Optional) A list of Host Groups the Volume is mapped to.
* `timeout` - (Optional) The number of seconds to wait to establish a connection the Silk server before returning a timeout error Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Volume. The convention is `silk-volume-timeString-hostID`
* `name` - The name of the Volume.
* `size_in_gb` - The size, in GB, of the Volume.
* `volume_group_name` - The name of the Volume Group that the Volume should be added to.
* `vmware` - This value corresponds to the 'VMware support' checkbox in the UI and specifies whether to enable VMFS. This value can be updated after the initial creation.
* `description` - A description of the Volume.
* `read_only` - This value corresponds to the 'Exposure Type' radio button in the UI and specifies whether the volume should be 'Read/Write' or 'Read Only'.
* `allow_destroy` - When set to true, this value will prevent the volume from being destroyed through Terraform.
* `host_mapping` - A list of Hosts the Volume is mapped to.
* `host_group_mapping` - A list of Host Groups the Volume is mapped to.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Volume from the Silk server. Before the volume can be destroyed, all mappings must be removed.
