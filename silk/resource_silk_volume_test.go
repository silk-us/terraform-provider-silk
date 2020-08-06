package silk

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

// Required Silk Centric Variables
var volumeName = "TerraformTestAccVolume"
var volumeGroupName = "TerraformTestAccVolumeGroup"
var hostNames = []string{"TerraformTestAccVolumeHost01", "TerraformTestAccVolumeHost02"}
var hostGroupNames = []string{"TerraformTestAccVolumeHG01", "TerraformTestAccVolumeHG02"}

// TestAccSilkVolume is the main function that is executed during the test process.
func TestAccSilkVolume(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckSilkCreateVolumeGroupPreCheck(volumeGroupName)
			testAccCheckSilkCreateHostsPreCheck(hostNames)
			testAccCheckSilkCreateHostGroupsPreCheck(hostGroupNames)

		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSilkVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSilkVolumeConfigBasic(volumeName, volumeGroupName, hostNames[0], hostGroupNames[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkVolumeExists("silk_volume.testacc"),
				),
			},
			{
				Config: testAccCheckSilkVolumeConfigAddMapping(volumeName, volumeGroupName, hostNames, hostGroupNames),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkVolumeExists("silk_volume.testacc"),
				),
			},
			{
				Config: testAccCheckSilkVolumeConfigRemoveMapping(volumeName, volumeGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkVolumeExists("silk_volume.testacc"),
				),
			},
		},
	})
}

// testAccCheckSilkCreateVolumeGroupPreCheck creates the Volume Group required for
// the full acceptance test
func testAccCheckSilkCreateVolumeGroupPreCheck(volumeGroupName string) error {

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	quotaInGb := 50
	enableDedup := true
	description := "Volume Group used for Terraform silk_volume Acceptance Testing"
	capacityPolicy := "default_vg_capacity_policy"

	_, err = silk.CreateVolumeGroup(volumeGroupName, quotaInGb, enableDedup, description, capacityPolicy)
	if err != nil {
		return err
	}

	return nil
}

// testAccCheckSilkCreateHostsPreCheck creates the Hosts required for
// the full acceptance test
func testAccCheckSilkCreateHostsPreCheck(hostNames []string) error {

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

// testAccCheckSilkCreateHostGroupsPreCheck creates the Host Groups required for
// the full acceptance test
func testAccCheckSilkCreateHostGroupsPreCheck(hostGroupNames []string) error {

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	description := "Host Group used for Terraform silk_volume Acceptance Testing"
	allowDifferentHostTypes := true

	for _, hostGroup := range hostGroupNames {
		_, err := silk.CreateHostGroup(hostGroup, description, allowDifferentHostTypes)
		if err != nil {
			return err
		}

	}

	return nil
}

// testAccCheckSilkVolumeConfigBasic returns a fully populated silk_volume resource
func testAccCheckSilkVolumeConfigBasic(name, volumeGroupName, hostName, hostGroupName string) string {
	return fmt.Sprintf(`
	resource "silk_volume" "testacc" {
		name = "%s"
		size_in_gb = 10
		volume_group_name = "%s"
		vmware = true
		description = "Volume used for Terraform silk_volume Acceptance Testing"
		read_only = false
		host_mapping = ["%s"]
		host_group_mapping = ["%s"]
		allow_destroy = true
	}
	`, name, volumeGroupName, hostName, hostGroupName)

}

// testAccCheckSilkVolumeConfigAddMapping adds additional host and host_groups to the previously created
// silk_volume resource
func testAccCheckSilkVolumeConfigAddMapping(name, volumeGroupName string, hostNames, hostGroupNames []string) string {
	return fmt.Sprintf(`
	resource "silk_volume" "testacc" {
		name = "%s"
		size_in_gb = 10
		volume_group_name = "%s"
		vmware = true
		description = "Volume used for Terraform silk_volume Acceptance Testing"
		read_only = false
		host_mapping = ["%s", "%s"]
		host_group_mapping = ["%s", "%s"]
		allow_destroy = true
	}
	`, name, volumeGroupName, hostNames[0], hostNames[1], hostGroupNames[0], hostGroupNames[1])

}

// testAccCheckSilkVolumeConfigAddMapping removes all of the host and host_groups to the previously created
// silk_volume resource. This validates the remove functionality while at the same time preparing the resource for
// terraform destroy (i.e you can't destroy the volume when it has mapping)
func testAccCheckSilkVolumeConfigRemoveMapping(name, volumeGroupName string) string {
	return fmt.Sprintf(`
	resource "silk_volume" "testacc" {
		name = "%s"
		size_in_gb = 10
		volume_group_name = "%s"
		vmware = true
		description = "Volume used for Terraform silk_volume Acceptance Testing"
		read_only = false
		host_mapping = []
		host_group_mapping = []
		allow_destroy = true
	}
	`, name, volumeGroupName)

}

// testAccCheckSilkVolumeConfigUpdate modifies the size_in_gb, vmware, description, and read_only
// paramaters to validated Update functionality
func testAccCheckSilkVolumeConfigUpdate(name, volumeGroupName string) string {
	return fmt.Sprintf(`
	resource "silk_volume" "testacc" {
		name = "%s"
		size_in_gb = 20
		volume_group_name = "%s"
		vmware = false
		description = "Updated. Volume used for Terraform silk_volume Acceptance Testing"
		read_only = true
		host_mapping = []
		host_group_mapping = []
		allow_destroy = true
	}
	`, name, volumeGroupName)

}

// testAccCheckSilkVolumeExists validates the resource was executed successfully
// by validating it exsits in the Terraform state
func testAccCheckSilkVolumeExists(n string) resource.TestCheckFunc {
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

// testAccCheckSilkVolumeDestroy deletes the Volume Group, Hosts, and Host Groups created
// as part of the PreCheck process
func testAccCheckSilkVolumeDestroy(s *terraform.State) error {

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	_, err = silk.DeleteVolumeGroup(volumeGroupName)
	if err != nil {
		return err
	}

	for _, host := range hostNames {
		_, err := silk.DeleteHost(host)
		if err != nil {
			return err
		}

	}

	for _, hostGroup := range hostGroupNames {
		_, err := silk.DeleteHostGroup(hostGroup)
		if err != nil {
			return err
		}

	}

	return nil
}
