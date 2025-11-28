package storageboxsnapshot_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/labelutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storagebox"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsnapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxSnapshotDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resStorageBox := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("snapshot-ds-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
		},
		Password: storagebox.GeneratePassword(t),
	}
	resStorageBox.SetRName("default")

	res := &storageboxsnapshot.RData{
		StorageBox:  resStorageBox.TFID() + ".id",
		Description: "tf-e2e-snapshot-ds",
		Labels: map[string]string{
			"key": randutil.GenerateID(),
		},
	}
	res.SetRName("default")

	byID := &storageboxsnapshot.DData{
		StorageBox: resStorageBox.TFID() + ".id",
		ID:         res.TFID() + ".id",
	}
	byID.SetRName("by_id")
	byName := &storageboxsnapshot.DData{
		StorageBox: resStorageBox.TFID() + ".id",
		Name:       res.TFID() + ".name",
	}
	byName.SetRName("by_name")
	byLabel := &storageboxsnapshot.DData{
		StorageBox:    resStorageBox.TFID() + ".id",
		LabelSelector: labelutil.Selector(res.Labels),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(storageboxsnapshot.ResourceType, storageboxsnapshot.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", res,

					"testdata/d/hcloud_storage_box_snapshot", byID,
					"testdata/d/hcloud_storage_box_snapshot", byName,
					"testdata/d/hcloud_storage_box_snapshot", byLabel,
				),
				ConfigStateChecks: slices.Concat(
					dataSourceAttributeStateCheck(res, byID.TFID(), tfjsonpath.Path{}),
					dataSourceAttributeStateCheck(res, byName.TFID(), tfjsonpath.Path{}),
					dataSourceAttributeStateCheck(res, byLabel.TFID(), tfjsonpath.Path{}),
				),
			},
		},
	})
}

func dataSourceAttributeStateCheck(res *storageboxsnapshot.RData, tfid string, path tfjsonpath.Path) []statecheck.StateCheck {
	return []statecheck.StateCheck{
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("storage_box_id"),
			tfid, path.AtMapKey("storage_box_id"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("id"),
			tfid, path.AtMapKey("id"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("name"),
			tfid, path.AtMapKey("name"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("description"),
			tfid, path.AtMapKey("description"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("is_automatic"),
			tfid, path.AtMapKey("is_automatic"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("labels"),
			tfid, path.AtMapKey("labels"),
			compare.ValuesSame(),
		),
		statecheck.ExpectKnownValue(tfid, path.AtMapKey("stats").AtMapKey("size"), knownvalue.NotNull()),
		statecheck.ExpectKnownValue(tfid, path.AtMapKey("stats").AtMapKey("size_filesystem"), knownvalue.NotNull()),
	}
}
