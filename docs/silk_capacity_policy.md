## silk_capacity_policy

Manage a Capacity Policy on the Silk Server.

## Example Usage

``` hcl
resource "silk_capacity_policy" "default" {
    name = "tf-cp-01"
    warningthreshold = 71
    errorthreshold = 75
    criticalthreshold = 90
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
* `snapshotoverheadthreshold` - (Optional) Percentage of capacity used by snapshots to generate an alert.

## Attribute Reference

The following attributes are exported:

* `criticalthreshold` - Percentage of used capacity required to trigger a 'critical' alert.
* `errorthreshold` - Percentage of used capacity required to trigger an 'error'.
* `id` - An ID unique to Terraform for this capacity Policy. The convention is `silk-capacityPolicy-capacityPolicyID-timeString`
* `name` - The name of the capacity policy.
* `snapshotoverheadthreshold` - Percentage of capacity used by snapshots to generate an alert.
* `warningthreshold` - Percentage of used capacity required to trigger a 'warning'.


## Destroy Behavior

On `terraform destroy`, this resource will remove the capacity Policy from the Silk server.

## Update Behavior

on `terraform update`, this resource will destroy and then create replacement. 
