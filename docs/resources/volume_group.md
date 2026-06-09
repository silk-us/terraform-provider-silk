---
page_title: "silk_volume_group Resource - terraform-provider-silk"
description: |-
  Manage a Volume Group on the Silk server.
---

# silk_volume_group (Resource)

Manage a Volume Group on the Silk server.

## Example Usage

```hcl
resource "silk_volume_group" "Silk-Volume-Group" {
  name        = "TerraformVolumeGroup"
  description = "Created through Terraform"
}
```

### Import

```
terraform import silk_volume_group.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Volume Group.
* `quota_in_gb` - (Optional) The size quota, in GB, of the Volume Group. The default of 0 corresponds to an Unlimited Quota.
* `enable_deduplication` - (Optional) This value corresponds to 'Provisioning Type' in the UI. When set to true, the Provisioning Type will be 'thin provisioning with dedupe'. Default is true.
* `description` - (Required) A description of the Volume Group.
* `capacity_policy` - (Optional) The capacity threshold policy profile for the Volume Group. Default is `default_vg_capacity_policy`.
* `timeout` - (Optional) The number of seconds to wait to establish a connection to the Silk server before returning a timeout error. Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Volume Group. The convention is `silk-volume-group-objID-timeString`.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Volume Group from the Silk server.
