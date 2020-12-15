package servertype

import (
	"fmt"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_server_type"
// template.
type DData struct {
	testtemplate.DataCommon

	ServerTypeID   string
	ServerTypeName string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// ServerTypesDData defines the fields for the "testdata/d/hcloud_server_types"
// template.
type ServerTypesDData struct {
	testtemplate.DataCommon
}

// TFID returns the data source identifier.
func (d *ServerTypesDData) TFID() string {
	return fmt.Sprintf("data.%s.%s", ServerTypesDataSourceType, d.RName())
}
