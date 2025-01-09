package placementgroup

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// ByID returns a function that obtains a placement_group by its ID.
func ByID(t *testing.T, placementGroup *hcloud.PlacementGroup) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.PlacementGroup.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find placement group %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if placementGroup != nil {
			*placementGroup = *found
		}
		return true
	}
}

// DData defines the fields for the "testdata/d/hcloud_placement_group" template.
type DData struct {
	testtemplate.DataCommon

	PlacementGroupID   string
	PlacementGroupName string
	LabelSelector      string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DData defines the fields for the "testdata/d/hcloud_placement_groups" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_placement_group"
// template.
type RData struct {
	testtemplate.DataCommon

	Name   string
	Labels map[string]string
	Type   string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

func NewRData(t *testing.T, name string, groupType string) *RData {
	rInt := acctest.RandInt()
	r := &RData{
		Name:   name,
		Type:   groupType,
		Labels: map[string]string{"key": strconv.Itoa(rInt)},
	}
	r.SetRName(name)
	return r
}
