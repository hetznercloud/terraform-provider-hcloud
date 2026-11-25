package floatingipassignement

import (
	"fmt"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

type RData struct {
	testtemplate.DataCommon

	FloatingIPID string
	ServerID     string
}

func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
