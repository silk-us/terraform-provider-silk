---
page_title: "silk_capacity_policy Resource - terraform-provider-silk"
description: |-
  Manage a Capacity Policy on the Silk server.
---

# silk_capacity_policy (Resource)

Manage a Capacity Policy on the Silk server.

## Example Usage

```hcl
resource "silk_capacity_policy" "default" {
  name                      = "tf-cp-01"
  warningthreshold          = 71
  errorthreshold            = 75
  criticalthreshold         = 90
  snapshotoverheadthreshold = 30
}
```

### Import

```
terraform import silk_capacity_policy.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Capacity Policy.
* `warningthreshold` - (Required) Percentage of used capacity required to trigger a 'warning'.
* `errorthreshold` - (Required) Percentage of used capacity required to trigger an 'error'.
* `criticalthreshold` - (Required) Percentage of used capacity required to trigger a 'critical' alert.
* `snapshotoverheadthreshold` - (Optional) Percentage of capacity used by snapshots to generate an alert. Set this to a value from 1 to 97. If it is omitted, `0` is sent, and the Silk server rejects `0`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Capacity Policy. The convention is `silk-capacityPolicy-capacityPolicyID-timeString`.

## Update Behavior

On `terraform apply` with changes, this resource will destroy and then create a replacement.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Capacity Policy from the Silk server.
