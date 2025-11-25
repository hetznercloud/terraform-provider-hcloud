package storageboxsnapshot_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storagebox"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsnapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxSnapshotResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	storageBox := &hcloud.StorageBox{}
	snapshot := &hcloud.StorageBoxSnapshot{}

	resStorageBox := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("storage-box-snapshot-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
		},
		Password: storagebox.GeneratePassword(t),
	}
	resStorageBox.SetRName("default")

	resMinimal := &storageboxsnapshot.RData{
		StorageBox: resStorageBox.TFID() + ".id",
	}
	resMinimal.SetRName("snapshot")

	resOptional := &storageboxsnapshot.RData{
		StorageBox:  resMinimal.StorageBox,
		Description: "tf-e2e-snapshot",
		Labels: map[string]string{
			"key": "value",
		},
	}
	resOptional.SetRName(resMinimal.RName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(storageboxsnapshot.ResourceType, storageboxsnapshot.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create with only required attributes

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", resMinimal,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resStorageBox.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resMinimal.TFID(), testsupport.CopyAPIResource(snapshot, storageboxsnapshot.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resMinimal.TFID(), tfjsonpath.New("storage_box_id"), testsupport.Int64ExactFromFunc(func() int64 { return storageBox.ID })),
					statecheck.ExpectKnownValue(resMinimal.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return snapshot.ID })),
					statecheck.ExpectKnownValue(resMinimal.TFID(), tfjsonpath.New("name"), testsupport.StringExactFromFunc(func() string { return snapshot.Name })),
					statecheck.ExpectKnownValue(resMinimal.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resMinimal.TFID(), tfjsonpath.New("is_automatic"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resMinimal.TFID(), tfjsonpath.New("labels"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				// Import

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", resMinimal,
				),
				ImportState:  true,
				ResourceName: resMinimal.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%d/%d", snapshot.StorageBox.ID, snapshot.ID), nil
				},
				ImportStateVerify: true,
			},
			{
				// Update with all optional attributes
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", resOptional,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(resOptional.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resStorageBox.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resOptional.TFID(), testsupport.CopyAPIResource(snapshot, storageboxsnapshot.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Same as before
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("storage_box_id"), testsupport.Int64ExactFromFunc(func() int64 { return storageBox.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return snapshot.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("name"), testsupport.StringExactFromFunc(func() string { return snapshot.Name })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("is_automatic"), knownvalue.Bool(false)),

					// Changed
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("tf-e2e-snapshot")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
				},
			},
			{
				// Create with all optional attributes
				Taint: []string{resOptional.TFID()}, // replace the resource
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_snapshot", resOptional,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resStorageBox.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resOptional.TFID(), testsupport.CopyAPIResource(snapshot, storageboxsnapshot.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("storage_box_id"), testsupport.Int64ExactFromFunc(func() int64 { return storageBox.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return snapshot.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("name"), testsupport.StringExactFromFunc(func() string { return snapshot.Name })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("is_automatic"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("tf-e2e-snapshot")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
				},
			},
		},
	})
}
