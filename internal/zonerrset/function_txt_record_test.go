package zonerrset_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
)

func TestTXTRecordFunction(t *testing.T) {
	t.Parallel()

	manyA := strings.Repeat("a", 255)
	someB := strings.Repeat("b", 10)

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::hcloud::txt_record("hello \"world\"")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact(`"hello \"world\""`)),
				},
			},
			{
				Config: fmt.Sprintf(`
				output "test" {
					value = provider::hcloud::txt_record("%s%s")
				}`, manyA, someB),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact(fmt.Sprintf(`"%s" "%s"`, manyA, someB))),
				},
			},
		},
	})
}
