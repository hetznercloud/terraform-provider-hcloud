package hcloud

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Deprecated: use testsupport.AccTestProviders instead
var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"hcloud": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

// Deprecated: use testsupport.AcceptanceTestPreCheck instead
func testAccHcloudPreCheck(t *testing.T) {
	if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
		t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
	}
}
