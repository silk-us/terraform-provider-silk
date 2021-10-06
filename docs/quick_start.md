# Quick Start Guide: Silk Terraform Provider

## Introduction to the Silk Terraform Provider

The Silk Provider exposes resources to interact with a Silk server.

## Installation

Requirements: Terraform has been successfully [installed](https://learn.hashicorp.com/terraform/getting-started/install.html).

1. Download the latest compiled binary from [GitHub releases](https://github.com/silk-us/silk-terraform-provider/releases).

   
``` 
macOS: terraform-provider-silk-darwin-amd64
Linux: terraform-provider-silk-linux-amd64
Windows: terraform-provider-silk-windows-amd64.exe
```

2. Move the Silk provider into the correct Terraform plugin directory

``` 
macOS: ~/.terraform.d/plugins/localdomain/provider/silk/1.0.9/darwin_amd64
Linux: ~/.terraform.d/plugins/localdomain/provider/silk/1.0.9/linux_amd64
Windows: %APPDATA%\terraform.d\plugins\localdomain\provider\silk\1.0.9\windows_amd64
   ```   
   _You may need to manually create the `plugin` directory._

3. Rename the the Silk provder to `terraform-provider-silk`
4. On Linux and macOS ensure that the binary has the appropriate permissions by running `chmod 744 terraform-provider-silk`
5. Run `terraform init` in the directory that contains your Terraform configuration fiile ( `main.tf` )

## Authentication

The Silk provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

* Environment variables
* Static credentials

### Environment variables

Storing credentials in environment variables is a more secure process than storing them in your source code, and it ensures that your credentials are not accidentally shared if your code is uploaded to an internal or public version control system such as GitHub. 

* **SILK_SDP_SERVER** (The IP Address of a Silk server.)
* **SILK_SDP_USERNAME** (The username used to authenticate against the Silk server.)
* **SILK_SDP_PASSWORD** (The password used to authenticate against the Silk Sever.)

``` hcl
terraform {
  required_providers {
    silk = {
      source  = "localdomain/provider/silk"
      version = "1.0.9"
    }
  }
}

provider "Silk" {}
```

#### Setting Environment Variables in Microsoft Windows

For Microsoft Windows-based operating systems, the environment variables can be set utilizing the setx command as follows:

``` 
setx SILK_SDP_SERVER "192.0.1.011"
setx SILK_SDP_USERNAME "admin"
setx SILK_SDP_PASSWORD "admin"
```

Run set without any other parameters to view current environment variables. Using setx saves the environment variables permanently, and the variables defined in the current shell will not be available until a new shell is opened. Using set instead of setx will define variables in the current shell session, but they will not be saved between sessions.

#### Setting Environment Variables in macOS and \*nix

For macOS and \*nix based operating systems the environment variables can be set utilizing the export command as follows:

``` 
export SILK_SDP_SERVER=192.0.1.011
export SILK_SDP_USERNAME=admin
export SILK_SDP_PASSWORD=admin
```

Run export without any other parameters to view current environment variables. In order for the environment variables to persist across terminal sessions, add the above three export commands to the `~\.bash_profile` or `~\.profile` file.

### Static credentials 

Static credentials can be provided by adding a `server`, `username` and `password` in-line in the
Silk provider block:

Usage:

``` hcl
provider "silk" {
  server     = "192.0.1.601"
  username    = "admin"
  password    = "admin"
}
```

## Sample Syntax

This section provides sample syntax to help you get started. For additional information and examples, see the [Silk Provider for Terraform Documentation](https://github.com/silk-us/silk-terraform-provider/tree/master/docs).

### Manage a Host

```hcl
provider "Silk" {}

resource "silk_host" "Silk-Host" {
  name = "TerraformHost"
  host_type = "Linux"
  pwwn = ["20:36:44:78:66:77:ab:10", "30:36:44:78:66:77:ab:10", "50:36:44:78:66:77:ab:10"]
}
```

## Silk Provider for Terraform Documentation

This guide acts only as a quick start to get up and running with the Silk Terraform Provider. For detailed information view the provided [Resource documentation](https://github.com/silk-us/silk-terraform-provider/tree/master/docs).
