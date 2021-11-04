## silk_host

Manage a Host on the Silk Server.

## Example Usage

``` hcl
resource "silk_host" "Silk-Host" {
  name = "TerraformHost"
  host_type = "Linux"
  pwwn = ["20:36:44:78:66:77:ab:10", "30:36:44:78:66:77:ab:10", "50:36:44:78:66:77:ab:10"]
}
```

### Import 

```
terraform import silk_host.{instance} {object name}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Host.
* `host_type` - (Required) The type of Host. Valid choices are "Linux", "Windows", and "ESX".
* `pwwn` - (Optional) An list of PWWNs that are mapped to the Host.
* `iqn` - (Optional) The IQN that is mapped to the Host.
* `timeout` - (Optional) The number of seconds to wait to establish a connection the Silk server before returning a timeout error Default is `15`.

## Attribute Reference

The following attributes are exported:

* `id` - An ID unique to Terraform for this Host. The convention is `silk-host-timeString-hostID`
* `name` - The name of the Host.
* `host_type` - The type of Host.
* `pwwn` - An list of PWWNs that are mapped to the Host.
* `iqn` - An list of IQNs that are mapped to the Host.

## Destroy Behavior

On `terraform destroy`, this resource will remove the Host from the Silk server.
