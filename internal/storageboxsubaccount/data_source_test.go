package storageboxsubaccount_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/labelutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storagebox"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsubaccount"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxSubaccountDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resStorageBox := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("subaccount-ds-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
		},
		Password: storagebox.GeneratePassword(t),
	}
	resStorageBox.SetRName("default")

	res := &storageboxsubaccount.RData{
		StorageBox:    resStorageBox.TFID() + ".id",
		HomeDirectory: "test",
		Password:      storagebox.GeneratePassword(t),
		Description:   "tf-e2e-subaccount-ds",
		Labels: map[string]string{
			"key": randutil.GenerateID(),
		},
		Raw: `
			access_settings = {
				reachable_externally = true
				samba_enabled = false
				ssh_enabled = true
				webdav_enabled = false
				readonly = true
			}`,
	}
	res.SetRName("default")

	byID := &storageboxsubaccount.DData{
		StorageBox: resStorageBox.TFID() + ".id",
		ID:         res.TFID() + ".id",
	}
	byID.SetRName("by_id")
	byUsername := &storageboxsubaccount.DData{
		StorageBox: resStorageBox.TFID() + ".id",
		Username:   res.TFID() + ".username",
	}
	byUsername.SetRName("by_username")
	byLabel := &storageboxsubaccount.DData{
		StorageBox:    resStorageBox.TFID() + ".id",
		LabelSelector: labelutil.Selector(res.Labels),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(storageboxsubaccount.ResourceType, storageboxsubaccount.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", res,

					"testdata/d/hcloud_storage_box_subaccount", byID,
					"testdata/d/hcloud_storage_box_subaccount", byUsername,
					"testdata/d/hcloud_storage_box_subaccount", byLabel,
				),
				ConfigStateChecks: slices.Concat(
					dataSourceAttributeStateCheck(res, byID.TFID(), tfjsonpath.Path{}),
					dataSourceAttributeStateCheck(res, byUsername.TFID(), tfjsonpath.Path{}),
					dataSourceAttributeStateCheck(res, byLabel.TFID(), tfjsonpath.Path{}),
				),
			},
		},
	})
}

func dataSourceAttributeStateCheck(res *storageboxsubaccount.RData, tfid string, path tfjsonpath.Path) []statecheck.StateCheck {
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
			res.TFID(), tfjsonpath.New("description"),
			tfid, path.AtMapKey("description"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("username"),
			tfid, path.AtMapKey("username"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("home_directory"),
			tfid, path.AtMapKey("home_directory"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("server"),
			tfid, path.AtMapKey("server"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("access_settings"),
			tfid, path.AtMapKey("access_settings"),
			compare.ValuesSame(),
		),
		statecheck.CompareValuePairs(
			res.TFID(), tfjsonpath.New("labels"),
			tfid, path.AtMapKey("labels"),
			compare.ValuesSame(),
		),
	}
}
