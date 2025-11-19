package storagebox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/labelutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storagebox"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res1 := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("list-1-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
			Labels:         map[string]string{"key": randutil.GenerateID()},
		},
		Password: storagebox.GeneratePassword(t),
	}
	res1.SetRName("main1")

	res2 := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("list-2-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
			Labels:         map[string]string{"key": randutil.GenerateID()},
		},
		Password: storagebox.GeneratePassword(t),
	}
	res2.SetRName("main2")

	all := &storagebox.DDataList{}
	all.SetRName("all")

	byLabel := &storagebox.DDataList{LabelSelector: labelutil.Selector(res1.Labels)}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", res1,
					"testdata/r/hcloud_storage_box", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", res1,
					"testdata/r/hcloud_storage_box", res2,
					"testdata/d/hcloud_storage_boxes", all,
					"testdata/d/hcloud_storage_boxes", byLabel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("storage_boxes"), knownvalue.SetSizeExact(1)),

					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("storage_boxes").AtSliceIndex(0).AtMapKey("name"), knownvalue.StringExact(res1.Name)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("storage_boxes").AtSliceIndex(0).AtMapKey("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxType)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("storage_boxes").AtSliceIndex(0).AtMapKey("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("storage_boxes").AtSliceIndex(0).AtMapKey("labels").AtMapKey("key"), knownvalue.StringExact(res1.Labels["key"])),
				},
			},
		},
	})
}
