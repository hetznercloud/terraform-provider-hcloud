package storageboxsnapshot_test

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsnapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxSnapshotDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resStorageBox := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("snapshot-ds-list-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
		},
		Password: storagebox.GeneratePassword(t),
	}
	resStorageBox.SetRName("default")

	res1 := &storageboxsnapshot.RData{
		StorageBox:  resStorageBox.TFID() + ".id",
		Description: "tf-e2e-snapshot-ds-1",
		Labels: map[string]string{
			"key": randutil.GenerateID(),
		},
	}
	res1.SetRName("default1")

	res2 := &storageboxsnapshot.RData{
		StorageBox:  resStorageBox.TFID() + ".id",
		Description: "tf-e2e-snapshot-ds-2",
		Labels: map[string]string{
			"key": randutil.GenerateID(),
		},
	}
	res2.SetRName("default2")

	all := &storageboxsnapshot.DDataList{
		StorageBox: resStorageBox.TFID() + ".id",
	}
	all.SetRName("all")

	byLabel := &storageboxsnapshot.DDataList{
		StorageBox:    resStorageBox.TFID() + ".id",
		LabelSelector: labelutil.Selector(res1.Labels),
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
					"testdata/r/hcloud_storage_box_snapshot", res1,
					"testdata/r/hcloud_storage_box_snapshot", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", res1,
					"testdata/r/hcloud_storage_box_snapshot", res2,

					"testdata/d/hcloud_storage_box_snapshots", all,
					"testdata/d/hcloud_storage_box_snapshots", byLabel,
				),
				ConfigStateChecks: slices.Concat(
					[]statecheck.StateCheck{
						// Making sure that multiple resources are returned
						statecheck.ExpectKnownValue(all.TFID(), tfjsonpath.New("snapshots"), knownvalue.SetExact(
							[]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{"labels": knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact(res1.Labels["key"])})}),
								knownvalue.ObjectPartial(map[string]knownvalue.Check{"labels": knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact(res2.Labels["key"])})}),
							}),
						),

						// Make sure label selector works
						statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("snapshots"), knownvalue.SetSizeExact(1)),
					},
					// Validate that all attributes are set correctly
					dataSourceAttributeStateCheck(res1, byLabel.TFID(), tfjsonpath.New("snapshots").AtSliceIndex(0)),
				),
			},
		},
	})
}
