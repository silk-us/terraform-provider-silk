package silk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/silk-us/silk-sdp-go-sdk/silksdp"
)

// TestAccSilkHost is the main function that is executed during the test process.
func TestAccSilkHost(t *testing.T) {

	// Required Silk Centric Variables.
	var hostName = "TerraformTestAccHost"
	var pwwns = []string{"20:21:22:23:45:67:89:ab", "30:11:12:23:45:67:89:ab"}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSilkHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSilkHostConfigBasic(hostName, pwwns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostExists("silk_host.testacc"),
				),
			},
			{
				Config: testAccCheckSilkHostConfigAddPWWN(hostName, pwwns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostExists("silk_host.testacc"),
				),
			},
			{
				Config: testAccCheckSilkHostConfigRemovePWWN(hostName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSilkHostExists("silk_host.testacc"),
				),
			},
		},
	})
}

// testAccCheckSilkHostConfigBasic returns a fully populated silk_host resource
func testAccCheckSilkHostConfigBasic(name string, pwwn []string) string {
	return fmt.Sprintf(`
	resource "silk_host" "testacc" {
		name = "%s"
		host_type = "Linux"
		pwwn = ["%s"]
	}
	`, name, pwwn[0])

}

// testAccCheckSilkHostConfigAddPWWN adds a PWWN to the previously created
// silk_host resource
func testAccCheckSilkHostConfigAddPWWN(name string, pwwn []string) string {
	return fmt.Sprintf(`
	resource "silk_host" "testacc" {
		name = "%s"
		host_type = "Linux"
		pwwn = ["%s", "%s"]
	}
	`, name, pwwn[0], pwwn[1])

}

// testAccCheckSilkHostConfigRemovePWWN removes all of the pwwns from the previously created
// silk_host resource
func testAccCheckSilkHostConfigRemovePWWN(name string) string {
	return fmt.Sprintf(`
	resource "silk_host" "testacc" {
		name = "%s"
		host_type = "Linux"
		pwwn = []
	}
	`, name)

}

// testAccCheckSilkHostExists validates the resource was executed successfully
// by validating it exsits in the Terraform state
func testAccCheckSilkHostExists(n string) resource.TestCheckFunc {
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

// testAccCheckSilkHostDestroy verifies the Host was sussuccessfully destroyed
// by the terraform destroy process
func testAccCheckSilkHostDestroy(s *terraform.State) error {

	// Required Silk Centric Variables
	var hostName = "TerraformTestAccHost"

	silk, err := silksdp.ConnectEnv()
	if err != nil {
		return err
	}

	// Validate the Volume has been destroyed
	_, err = silk.GetHostID(hostName)
	if err != nil {
		if strings.Contains(err.Error(), "The server does not contain") {
			return nil
		}
		return err

	}

	return nil
}
