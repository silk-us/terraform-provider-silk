# Contributing

Welcome to the Silk Terraform Provider. This document provides an overview of the Provider Development process and how to contribute to the repository.

## Development Environment

The [.devcontainer](https://github.com/silk-us/silk-terraform-provider/tree/master/.devcontainer) contains a preconfigured `Dockerfile` that was designed to provide all the necessary 
environment configurations required to develop the Silk Terraform Provider. This mainly includes the installation of GoLang and Terraform along with various associated tools. The simpliest way of accessing this Docker container is through the [VS Code Remote Container](https://code.visualstudio.com/docs/remote/containers) feature which has also been configured through the `devcontainer.json` file which automatically installs the VS Code extensions that help simplify the development process.

### GoLang Requirements

All GoLang requirements for the Provider are managead through [Go Modules](https://blog.golang.org/using-go-modules). More specifically this includes the [go.mod](https://github.com/silk-us/silk-terraform-provider/blob/master/go.mod) and `go.sum` files. The `go.sum` file is automatically generated and should never need to be manually managed. The `go.mod` is also automatically managed through the `go mod vendor` command. This means that any new modules that are added to the Provider code will be added to the `go.mod` file after running the `go mod vendor` command.

To update a required package -- which should usually just be the Silk SDP Go SDK -- you will need to update the version listed in the `go.mod` file and then run the `go mod vendor` command which will automatically download the new version to the `vendor` folder which will be included in the packaged Terraform binary.

### Build a new Development Terraform Provider

The Terraform Build workflow has been defined in the [Makefile](https://github.com/silk-us/silk-terraform-provider/blob/master/Makefile). To access the Build workflow run `make build` which will create a new version of the Provider and then run `terraform init`.

Once built, you do not need to install the new binary since Terraform will automatically detect it in the root directory.

### Release

The Provider release workflow has been defined in the [Makefile](https://github.com/silk-us/silk-terraform-provider/blob/master/Makefile). To access the Release workflow run `make release` which will create a new version of the Provider, for each supported operating system, in the `./bin` directory. Once created, these files should be uploaded to the [GitHub Releases page](https://github.com/silk-us/silk-terraform-provider/releases).

## Acceptance Test

Each Resource includes an Acceptance Test which "use real Terraform configurations to exercise the code in real plan, apply, refresh, and destroy life cycles." The Acceptance Test  workflow has been defined in the [Makefile](https://github.com/silk-us/silk-terraform-provider/blob/master/Makefile). To execute the acceptance tests run `make testacc`. 

The Acceptance Tests will create real resources in a provided Silk platform. This means that the `SILK_SDP_SERVER`, `SILK_SDP_USERNAME`, `SILK_SDP_PASSWORD` will need to be configured before running the Acceptance Tests. It's also important to note that there if an Acceptance Test fails the environment clean process may not be successfully executed so a manual cleanup may be required in this case.
