---
page_title: "silk_volume_group_view Resource - terraform-provider-silk"
description: |-
  Manage a Volume Group View on the Silk server.
---

# silk_volume_group_view (Resource)

Manage a Volume Group View on the Silk server. A view is a writable, exposable copy of a [`silk_volume_group_snapshot`](volume_group_snapshot.md).

Views created by Terraform are always created with `is_auto_deleteable = false`, so the Retention Policy will never expire a view that Terraform is tracking.

## Example Usage

```hcl
resource "silk_volume_group_view" "example" {
  name             = "view01"
  snapshot_name    = silk_volume_group_snapshot.example.full_name
  retention_policy = silk_retention_policy.example.name
}
```

### Import

Import with the full name the Silk server assigns, `{volume group name}:{view name}`.

```
terraform import silk_volume_group_view.{instance} {volume group name}:{view name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The short name of the view. The Silk server names the view `{volume group name}:{name}`, using the Volume Group of the source snapshot. Changing this value forces a new View to be created.
* `snapshot_name` - (Required) The snapshot to create the view from. Accepts the full `{volume group name}:{snapshot name}` or the short snapshot name. The short name only works when one Volume Group has a snapshot with that name. Changing this value forces a new View to be created.
* `retention_policy` - (Required) The name of the Retention Policy attached to the view. Changing this value forces a new View to be created.
* `timeout` - (Optional) The number of seconds to wait to establish a connection to the Silk server before returning a timeout error. Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this View. The convention is `silk-view-objID-timeString`.
* `obj_id` - The SDP ID of the View.
* `full_name` - The full name the Silk server assigns, `{volume group name}:{name}`.
* `volume_group_name` - The Volume Group the view belongs to.
* `is_auto_deleteable` - Always `false` when created by Terraform. If it is changed to `true` outside of Terraform, the next plan shows a warning. Taint and re-apply the resource to replace it.

## Update Behavior

The Silk server does not support modifying views. Any change other than `timeout` destroys the View and creates a replacement.

## Destroy Behavior

On `terraform destroy`, this resource will remove the View from the Silk server. A view with active Host mappings must have those mappings removed first.
