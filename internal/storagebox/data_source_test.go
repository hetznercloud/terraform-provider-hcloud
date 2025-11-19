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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("datasource-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
			Labels:         map[string]string{"key": randutil.GenerateID()},
		},
		Password: generatePassword(t),
	}
	res.SetRName("main")

	byID := &storagebox.DData{
		ID: res.TFID() + ".id",
	}
	byID.SetRName("by_id")
	byName := &storagebox.DData{
		Name: res.TFID() + ".name",
	}
	byName.SetRName("by_name")
	byLabel := &storagebox.DData{
		LabelSelector: labelutil.Selector(res.Labels),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(storagebox.ResourceType, storagebox.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", res,
					"testdata/d/hcloud_storage_box", byID,
					"testdata/d/hcloud_storage_box", byName,
					"testdata/d/hcloud_storage_box", byLabel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(res.Name)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxType)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("labels").AtMapKey("key"), knownvalue.StringExact(res.Labels["key"])),

					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(res.Name)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxType)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("labels").AtMapKey("key"), knownvalue.StringExact(res.Labels["key"])),

					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(res.Name)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxType)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("labels").AtMapKey("key"), knownvalue.StringExact(res.Labels["key"])),
				},
			},
		},
	})
}
