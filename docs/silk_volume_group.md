## silk_volume_group

Manage a Volume Group on the Silk Server.

## Example Usage

``` hcl
resource "silk_volume_group" "Silk-Volume-Group" {
  name = "TerraformVolumeGroup"
  description = "Crated through TF"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Volume Group.
* `quota_in_gb` - (Optional) The size quota, in GB, of the Volume Group. The Default option of 0 corresponds to an Unlimited Quota.
* `enable_deduplication` - (Optional) This value corresponds to 'Provisioning Type' in the UI. When set to true, the Provisioning Type will be 'thin provisioning with dedupe'. Default value is true
* `description` - (Required) A description of the Volume Group
* `capacity_policy` - (Optional) The capacity threshold policy profile for the Volume Group. Default is default_vg_capacity_policy.
* `timeout` - (Optional) The number of seconds to wait to establish a connection the Silk server before returning a timeout error Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Volume Group. The convention is `silk-volume-group-timeString-hostID`
* `name` - The name of the Volume Group.
* `quota_in_gb` - The size quota, in GB, of the Volume Group.
* `enable_deduplication` - This value corresponds to 'Provisioning Type' in the UI. When set to true, the Provisioning Type will be 'thin provisioning with dedupe'.
* `description` - A description of the Volume Group
* `capacity_policy` - The capacity threshold policy profile for the Volume Group.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Volume Group from the Silk server.
