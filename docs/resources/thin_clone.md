---
page_title: "silk_thin_clone Resource - terraform-provider-silk"
description: |-
  Manage a Thin Clone on the Silk server.
---

# silk_thin_clone (Resource)

Manage a Thin Clone on the Silk server. A Thin Clone is a writable Volume that is backed by a Volume Group Snapshot. The resulting object lives at `/volumes` on the Silk server, so it can be queried, mapped, and updated using the same surface as a regular Volume.

## Example Usage

```hcl
resource "silk_thin_clone" "Silk-Thin-Clone" {
  name                 = "ExampleThinCloneName"
  volume_group_name    = "ExampleVolumeGroupName"
  source_snapshot_name = "ExampleVolumeGroupSnapshotName"
  source_volume_name   = "ExampleSourceVolumeName"
  description          = "Created through Terraform"
  read_only            = false
  host_mapping         = ["ExampleHostName"]
  host_group_mapping   = ["ExampleHostGroupName"]
  allow_destroy        = true
}
```

### Import

```
terraform import silk_thin_clone.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Thin Clone.
* `volume_group_name` - (Required) The name of the Volume Group the Thin Clone is created in. The Volume Group must contain the source Volume.
* `source_snapshot_name` - (Required) The name of the Volume Group Snapshot to clone from. Changing this value forces a new Thin Clone to be created.
* `source_volume_name` - (Required) The name of the source Volume within the snapshot to clone. Changing this value forces a new Thin Clone to be created.
* `description` - (Optional) A description of the Thin Clone. If omitted, the Silk server populates this with `Clone of volume <name>`.
* `read_only` - (Optional) This value corresponds to the 'Exposure Type' radio button in the UI and specifies whether the Thin Clone should be 'Read/Write' or 'Read Only'. Default is false.
* `allow_destroy` - (Optional) When set to true, this value will allow the Thin Clone to be destroyed through Terraform. Default is false.
* `host_mapping` - (Optional) A list of Hosts the Thin Clone is mapped to.
* `host_group_mapping` - (Optional) A list of Host Groups the Thin Clone is mapped to.
* `timeout` - (Optional) The number of seconds to wait to establish a connection to the Silk server before returning a timeout error. Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Thin Clone. The convention is `silk-thin-clone-objID-timeString`.
* `obj_id` - The SDP ID of the Thin Clone (lives in `/volumes`).
* `volume_group_id` - The SDP ID of the Volume Group the Thin Clone belongs to. Used to detect Volume Group renames.
* `source_snap_id` - The SDP ID of the Volume Group Snapshot that backs this Thin Clone.
* `size_in_gb` - The size, in GB, of the Thin Clone. Inherited from the source Volume.
* `vmware` - VMware support flag. Inherited from the source Volume.
* `scsi_sn` - The SCSI serial number as a string.

## Destroy Behavior

On `terraform destroy`, this resource will remove all Host and Host Group mappings from the Thin Clone and then remove the Thin Clone (Volume) from the Silk server. The `allow_destroy` argument must be set to `true` for the destroy to succeed; otherwise Terraform will return an error and leave the Thin Clone in place.
