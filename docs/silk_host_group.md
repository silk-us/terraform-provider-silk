# silk_host_group

Manage a Host Group on the Silk Server.

## Example Usage

``` hcl
resource "silk_host_group" "Silk-Host-Group" {
  name = "TerraformHostGroup"
  description = "Created through Terraform"
  allow_different_host_types = true
  host_mapping = ["ExampleHostName", "ExampleHostName"]
}
```

### Import 

```
terraform import silk_host_group.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Host Group.
* `description` - (Required) A description of the Host Group
* `allow_different_host_types` - (Optional) Corresponds to the 'Enable mixed host OS types' checkbox in the UI. The default value is false.
* `host_mapping` - (Optional) A list of Hosts that belong to the Host Group.
* `timeout` - (Optional) The number of seconds to wait to establish a connection the Silk server before returning a timeout error Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Host Group. The convention is `silk-host-group-timeString-hostID`
* `name` - The name of the Host Group.
* `description` - A description of the Host Group
* `allow_different_host_types` - Corresponds to the 'Enable mixed host OS types' checkbox in the UI. The default value is false.
* `host_mapping` - A list of Hosts that belong to the Host Group.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Host Group from the Silk server. All Hosts must be removed from the Host Group before the destroy will succeed.
