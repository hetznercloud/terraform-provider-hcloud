package networkmember

import (
	"fmt"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

type DDataList struct {
	testtemplate.DataCommon

	NetworkID string

	WithType   []string
	WithStatus []string
	WithSubnet []string

	Raw string
}

func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}
