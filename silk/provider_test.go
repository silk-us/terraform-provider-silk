package silk

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"silk": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()

}

func testAccPreCheck(t *testing.T) {

	if err := os.Getenv("SILK_SDP_SERVER"); err == "" {
		t.Fatal("SILK_SDP_SERVER must be set for acceptance tests")
	}
	if err := os.Getenv("SILK_SDP_USERNAME"); err == "" {
		t.Fatal("SILK_SDP_USERNAME must be set for acceptance tests")
	}
	if err := os.Getenv("SILK_SDP_PASSWORD"); err == "" {
		t.Fatal("SILK_SDP_PASSWORD must be set for acceptance tests")
	}
}
