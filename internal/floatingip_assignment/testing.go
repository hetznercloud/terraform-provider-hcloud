package floatingip_assignment

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

// RData defines the fields for the "testdata/r/hcloud_floating_ip_assignment" template.
type RData struct {
	testtemplate.DataCommon

	FloatingIPID string
	ServerID     string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
