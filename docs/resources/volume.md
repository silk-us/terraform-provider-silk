---
page_title: "silk_volume Resource - terraform-provider-silk"
description: |-
  Manage a Volume on the Silk server.
---

# silk_volume (Resource)

Manage a Volume on the Silk server.

## Example Usage

```hcl
resource "silk_volume" "Silk-Volume" {
  name               = "ExampleVolumeName"
  size_in_gb         = 10
  volume_group_name  = "ExampleVolumeGroupName"
  vmware             = true
  description        = "Created through Terraform"
  read_only          = false
  host_mapping       = ["ExampleHostName"]
  host_group_mapping = ["ExampleHostGroupName"]
  allow_destroy      = true
}
```

### Import

```
terraform import silk_volume.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Volume.
* `size_in_gb` - (Required) The size, in GB, of the Volume.
* `volume_group_name` - (Required) The name of the Volume Group that the Volume should be added to.
* `vmware` - (Optional) This value corresponds to the 'VMware support' checkbox in the UI and specifies whether to enable VMFS. This value can be updated after the initial creation. Default is false.
* `description` - (Required) A description of the Volume.
* `read_only` - (Optional) This value corresponds to the 'Exposure Type' radio button in the UI and specifies whether the volume should be 'Read/Write' or 'Read Only'. Default is false.
* `allow_destroy` - (Optional) When set to true, this value will allow the volume to be destroyed through Terraform. Default is false.
* `host_mapping` - (Optional) A list of Hosts the Volume is mapped to.
* `host_group_mapping` - (Optional) A list of Host Groups the Volume is mapped to.
* `timeout` - (Optional) The number of seconds to wait to establish a connection to the Silk server before returning a timeout error. Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Volume. The convention is `silk-volume-objID-timeString`.
* `obj_id` - The SDP ID of the Volume.
* `scsi_sn` - The SCSI serial number as a string.

## Destroy Behavior

On `terraform destroy`, this resource will remove all Host and Host Group mappings from the Volume and then remove the Volume from the Silk server. The `allow_destroy` argument must be set to `true` for the destroy to succeed.
