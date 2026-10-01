---
page_title: "silk_volume_group_snapshot Resource - terraform-provider-silk"
description: |-
  Manage a Volume Group Snapshot on the Silk server.
---

# silk_volume_group_snapshot (Resource)

Manage a Volume Group Snapshot on the Silk server. A snapshot is a point-in-time, read-only copy of every Volume in a Volume Group. Snapshots are the source for [`silk_volume_group_view`](volume_group_view.md) and [`silk_thin_clone`](thin_clone.md).

Snapshots created by Terraform are always created with `is_auto_deleteable = false`, so the Retention Policy will never expire a snapshot that Terraform is tracking.

## Example Usage

```hcl
resource "silk_retention_policy" "example" {
  name          = "ExampleRetentionPolicy"
  num_snapshots = "5"
  weeks         = "0"
  days          = "1"
  hours         = "0"
}

resource "silk_volume_group_snapshot" "example" {
  name              = "snap01"
  volume_group_name = silk_volume_group.example.name
  retention_policy  = silk_retention_policy.example.name

  # take the snapshot after the volumes exist
  depends_on = [silk_volume.example]
}
```

### Import

Import with the full name the Silk server assigns, `{volume group name}:{snapshot name}`.

```
terraform import silk_volume_group_snapshot.{instance} {volume group name}:{snapshot name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The short name of the snapshot. The Silk server names the snapshot `{volume_group_name}:{name}`. Changing this value forces a new Snapshot to be created.
* `volume_group_name` - (Required) The name of the Volume Group to snapshot. Changing this value forces a new Snapshot to be created.
* `retention_policy` - (Required) The name of the Retention Policy attached to the snapshot. Changing this value forces a new Snapshot to be created.
* `timeout` - (Optional) The number of seconds to wait to establish a connection to the Silk server before returning a timeout error. Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Snapshot. The convention is `silk-snapshot-objID-timeString`.
* `obj_id` - The SDP ID of the Snapshot.
* `full_name` - The full name the Silk server assigns, `{volume_group_name}:{name}`. Use this when referencing the snapshot from a [`silk_volume_group_view`](volume_group_view.md).
* `is_auto_deleteable` - Always `false` when created by Terraform. If it is changed to `true` outside of Terraform, the next plan shows a warning. Taint and re-apply the resource to replace it.

## Update Behavior

The Silk server does not support modifying snapshots. Any change other than `timeout` destroys the Snapshot and creates a replacement. Views and Thin Clones created from the Snapshot are replaced with it.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Snapshot from the Silk server. Views and Thin Clones created from the Snapshot must be removed first. Terraform handles this ordering when they reference the Snapshot.
