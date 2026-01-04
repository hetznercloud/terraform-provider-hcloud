package storageboxsubaccount_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
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

func TestAccStorageBoxSubaccountDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resStorageBox := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("subaccount-ds-list-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
		},
		Password: storagebox.GeneratePassword(t),
	}
	resStorageBox.SetRName("default")

	res1 := &storageboxsubaccount.RData{
		StorageBox:    resStorageBox.TFID() + ".id",
		HomeDirectory: "test-1",
		Password:      storagebox.GeneratePassword(t),
		Description:   "tf-e2e-subaccount-ds-1",
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
	res1.SetRName("default1")

	res2 := &storageboxsubaccount.RData{
		StorageBox:    resStorageBox.TFID() + ".id",
		HomeDirectory: "test-2",
		Password:      storagebox.GeneratePassword(t),
		Description:   "tf-e2e-subaccount-ds-2",
		Labels: map[string]string{
			"key": randutil.GenerateID(),
		},
		Raw: `
			access_settings = {
				reachable_externally = false
				samba_enabled = true
				ssh_enabled = false
				webdav_enabled = true
				readonly = false
			}
		`,
	}
	res2.SetRName("default2")

	all := &storageboxsubaccount.DDataList{
		StorageBox: resStorageBox.TFID() + ".id",
	}
	all.SetRName("all")

	byLabel := &storageboxsubaccount.DDataList{
		StorageBox:    resStorageBox.TFID() + ".id",
		LabelSelector: labelutil.Selector(res1.Labels),
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
					"testdata/r/hcloud_storage_box_subaccount", res1,
					"testdata/r/hcloud_storage_box_subaccount", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", res1,
					"testdata/r/hcloud_storage_box_subaccount", res2,

					"testdata/d/hcloud_storage_box_subaccounts", all,
					"testdata/d/hcloud_storage_box_subaccounts", byLabel,
				),
				ConfigStateChecks: slices.Concat(
					[]statecheck.StateCheck{
						// Making sure that multiple resources are returned
						statecheck.ExpectKnownValue(all.TFID(), tfjsonpath.New("subaccounts"), knownvalue.SetExact(
							[]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{"labels": knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact(res1.Labels["key"])})}),
								knownvalue.ObjectPartial(map[string]knownvalue.Check{"labels": knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact(res2.Labels["key"])})}),
							}),
						),

						// Make sure label selector works
						statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("subaccounts"), knownvalue.SetSizeExact(1)),
					},
					// Validate that all attributes are set correctly
					dataSourceAttributeStateCheck(res1, byLabel.TFID(), tfjsonpath.New("subaccounts").AtSliceIndex(0)),
				),
			},
		},
	})
}
