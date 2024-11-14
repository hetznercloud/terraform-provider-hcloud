package loadbalancertype

import (
	"fmt"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_load_balancer_type"
// template.
type DData struct {
	testtemplate.DataCommon

	LoadBalancerTypeID   string
	LoadBalancerTypeName string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", loadbalancertype.DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_load_balancer_types"
// template.
type DDataList struct {
	testtemplate.DataCommon
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", loadbalancertype.DataSourceListType, d.RName())
}
