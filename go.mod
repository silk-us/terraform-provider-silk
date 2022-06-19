module github.com/silk-us/silk-terraform-provider

go 1.14

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0
	github.com/silk-us/silk-sdp-go-sdk v1.1.1
)

replace (
	// github.com/silk-us/silk-sdp-go-sdk => /mnt/d/Dropbox/SDP-Terraform/PC/silk-sdp-go-sdk
	// github.com/silk-us/silk-sdp-go-sdk => ../silk-sdp-go-sdk
)
