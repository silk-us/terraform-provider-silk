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
func TestAccSilkHostGroup(t *testing.T) {

	// Required Silk Centric Variables.
	var hostGroupName = "TerraformTestAccHostGroup"
	var hostNames = []string{"TerraformTestAccHGHost01", "TerraformTestAccHGHost02"}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckSilkHostGroupCreateHostsPreCheck(hostNames)

		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSilkHostGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSilkHostGroupConfigBasic(hostGroupName, hostNames[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostGroupExists("silk_host_group.testacc"),
				),
			},
			{
				Config: testAccCheckSilkHostGroupConfigAddMapping(hostGroupName, hostNames),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostGroupExists("silk_host_group.testacc"),
				),
			},
			{
				Config: testAccCheckSilkHostGroupConfigRemoveMapping(hostGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostGroupExists("silk_host_group.testacc"),
				),
			},
			{
				Config: testAccCheckSilkHostGroupConfigUpdate(hostGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostGroupExists("silk_host_group.testacc"),
				),
			},
		},
	})
}

// testAccCheckSilkCreateHostsPreCheck creates the Hosts required for
// the full acceptance test
func testAccCheckSilkHostGroupCreateHostsPreCheck(hostNames []string) error {

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	hostType := "Linux"

	for _, host := range hostNames {
		_, err := silk.CreateHost(host, hostType)
		if err != nil {
			return err
		}

	}

	return nil
}

// testAccCheckSilkHostGroupConfigBasic returns a fully populated silk_host_group resource
func testAccCheckSilkHostGroupConfigBasic(name, hostName string) string {
	return fmt.Sprintf(`
	resource "silk_host_group" "testacc" {
		name = "%s"
		description = "Host Group used for Terraform silk_host_group Acceptance Testing"
		allow_different_host_types = false
		host_mapping = ["%s"]
	}
	`, name, hostName)

}

// testAccCheckSilkHostGroupConfigAddMapping adds additional host and host_groups to the previously created
// silk_host_group resource
func testAccCheckSilkHostGroupConfigAddMapping(name string, hostNames []string) string {
	return fmt.Sprintf(`
	resource "silk_host_group" "testacc" {
		name = "%s"
		description = "Host Group used for Terraform silk_host_group Acceptance Testing"
		allow_different_host_types = false
		host_mapping = ["%s", "%s"]
	}
	`, name, hostNames[0], hostNames[1])

}

// testAccCheckSilkHostGroupConfigAddMapping removes all of the host and host_groups to the previously created
// silk_host_group resource. This validates the remove functionality while at the same time preparing the resource for
// terraform destroy (i.e you can't destroy the volume when it has mapping)
func testAccCheckSilkHostGroupConfigRemoveMapping(name string) string {
	return fmt.Sprintf(`
	resource "silk_host_group" "testacc" {
		name = "%s"
		description = "Host Group used for Terraform silk_host_group Acceptance Testing"
		allow_different_host_types = false
		host_mapping = []
	}
	`, name)

}

// testAccCheckSilkHostGroupConfigUpdate modifies the size_in_gb, vmware, description, and read_only
// paramaters to validated Update functionality
func testAccCheckSilkHostGroupConfigUpdate(name string) string {
	return fmt.Sprintf(`
	resource "silk_host_group" "testacc" {
		name = "%s"
		description = "Updated. Host Group used for Terraform silk_host_group Acceptance Testing"
		allow_different_host_types = true
		host_mapping = []
	}
	`, name)

}

// testAccCheckSilkHostGroupExists validates the resource was executed successfully
// by validating it exsits in the Terraform state
func testAccCheckSilkHostGroupExists(n string) resource.TestCheckFunc {
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

// testAccCheckSilkHostGroupDestroy deletes the Volume Group, Hosts, and Host Groups created
// as part of the PreCheck process and then verifies the Volume was sussuccessfully destroyed
// by the terraform destroy process
func testAccCheckSilkHostGroupDestroy(s *terraform.State) error {

	// Required Silk Centric Variables
	var hostGroupName = "TerraformTestAccHostGroup"

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	// Validate the Volume has been destroyed
	_, err = silk.GetHostGroupID(hostGroupName)
	if err != nil {
		if strings.Contains(err.Error(), "The server does not contain") {
			return nil
		}
		return err

	}

	return nil
}
