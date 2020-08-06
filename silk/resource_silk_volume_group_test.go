package silk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

// TestAccSilkVolume is the main function that is executed during the test process.
func TestAccSilkVolumeGroup(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSilkVolumeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSilkVolumeGroupConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkVolumeGroupExists("silk_volume_group.testacc"),
				),
			},
			{
				Config: testAccCheckSilkVolumeGroupConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkVolumeGroupExists("silk_volume_group.testacc"),
				),
			},
		},
	})
}

// testAccCheckSilkVolumeGroupConfigBasic returns a fully populated silk_volume_group resource
func testAccCheckSilkVolumeGroupConfigBasic() string {
	return fmt.Sprintf(`
	resource "silk_volume_group" "testacc" {
		name = "TerraformTestAccVolumeGroup"
		quota_in_gb = 10
		enable_deduplication = true
		description = "Volume used for Terraform silk_volume_group Acceptance Testing"
	}
	`)

}

// testAccCheckSilkVolumeGroupConfigUpdate modifies the name, quota_in_gb, and description
// paramaters to validated Update functionality
func testAccCheckSilkVolumeGroupConfigUpdate() string {
	return fmt.Sprintf(`
	resource "silk_volume_group" "testacc" {
		name = "TerraformTestAccVolumeGroupNew"
		quota_in_gb = 20
		enable_deduplication = true
		description = "Updated. Volume used for Terraform silk_volume_group Acceptance Testing"
	}
	`)

}

// testAccCheckSilkVolumeGroupExists validates the resource was executed successfully
// by validating it exsits in the Terraform state
func testAccCheckSilkVolumeGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if resources.Primary.ID == "" {
			return fmt.Errorf("No Volume set")
		}

		return nil
	}
}

// testAccCheckSilkVolumeGroupDestroy deletes the Volume Group, Hosts, and Host Groups created
// as part of the PreCheck process
func testAccCheckSilkVolumeGroupDestroy(s *terraform.State) error {

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	// Validate the Volume Group has been destroyed.
	_, err = silk.GetVolumeGroupID("TerraformTestAccVolumeUpdated")
	if err != nil {
		if strings.Contains(err.Error(), "The server does not contain") {
			return nil
		}
		return err
	}

	return nil
}
