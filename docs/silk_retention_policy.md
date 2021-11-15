## silk_retention_policy

Manage a Retention Policy on the Silk Server.

## Example Usage

``` hcl
resource "silk_retention_policy" "default" {
    name = "Weekly Retention"
    num_snapshots = "7"
    weeks = "1"
    days = "0"
    hours = "0"
}
```

### Import 

```
terraform import silk_retention_policy.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Retention Policy.
* `num_snapshots` - (Required) The total number of snapshots this policy can hold.
* `weeks` - (Required) The number of weeks to retain the snapshot.
* `days` - (Required) The number of days to retain the snapshot.
* `hours` - (Required) The number of hours to retain the snapshot.

## Attribute Reference

The following attributes are exported:

* `days` - The number of days to retain the snapshot.
* `hours` - The number of hours to retain the snapshot.
* `id` - An ID unique to Terraform for this Retention Policy. The convention is `silk-RetentionPolicy-retentionPolicyID-timeString`
* `name` - The name of the retention policy.
* `num_snapshots` - The type of Host.
* `weeks` - The number of weeks to retain the snapshot.


## Destroy Behavior

On `terraform destroy`, this resource will remove the Retention Policy from the Silk server.
