package storageboxsubaccount_test

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsubaccount"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxSubaccountResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	storageBox := &hcloud.StorageBox{}
	subaccount := &hcloud.StorageBoxSubaccount{}

	resStorageBox := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("storage-box-subaccount-%s", randutil.GenerateID()),
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
	}
	res.SetRName("subaccount")

	resOptional := testtemplate.DeepCopy(t, res)
	resOptional.HomeDirectory = "updated"
	resOptional.Password = storagebox.GeneratePassword(t)
	resOptional.Description = "tf-e2e-subaccount"
	resOptional.Labels = map[string]string{
		"key": "value",
	}
	resOptional.Raw = `
		access_settings = {
			reachable_externally = true
			samba_enabled = false
			ssh_enabled = true
			webdav_enabled = false
			readonly = true
		}
	`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(storageboxsubaccount.ResourceType, storageboxsubaccount.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create with only required attributes

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resStorageBox.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(res.TFID(), testsupport.CopyAPIResource(subaccount, storageboxsubaccount.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("storage_box_id"), testsupport.Int64ExactFromFunc(func() int64 { return storageBox.ID })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return subaccount.ID })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("username"), testsupport.StringExactFromFunc(func() string { return subaccount.Username })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("server"), testsupport.StringExactFromFunc(func() string { return subaccount.Server })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("home_directory"), knownvalue.StringExact("test")),
					statecheck.ExpectSensitiveValue(res.TFID(), tfjsonpath.New("password")),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("password"), knownvalue.StringExact(res.Password)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("reachable_externally"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("samba_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("ssh_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("webdav_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("readonly"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("labels"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				// Import

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", res,
				),
				ImportState:  true,
				ResourceName: res.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%d/%d", subaccount.StorageBox.ID, subaccount.ID), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // Not returned in the API
			},
			{
				// Update with all optional attributes

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", resOptional,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(resOptional.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resStorageBox.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resOptional.TFID(), testsupport.CopyAPIResource(subaccount, storageboxsubaccount.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Same as before
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("storage_box_id"), testsupport.Int64ExactFromFunc(func() int64 { return storageBox.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return subaccount.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("username"), testsupport.StringExactFromFunc(func() string { return subaccount.Username })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("server"), testsupport.StringExactFromFunc(func() string { return subaccount.Server })),

					// Changed (or will be changed in Actions PR)
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("tf-e2e-subaccount")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("home_directory"), knownvalue.StringExact("updated")),
					statecheck.ExpectSensitiveValue(resOptional.TFID(), tfjsonpath.New("password")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("password"), knownvalue.StringExact(resOptional.Password)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("reachable_externally"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("samba_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("ssh_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("webdav_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("readonly"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
				},
			},
			{
				// Create with all optional attributes

				Taint: []string{resOptional.TFID()}, // replace the resource
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_storage_box", resStorageBox,
					"testdata/r/hcloud_storage_box_subaccount", resOptional,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resStorageBox.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resOptional.TFID(), testsupport.CopyAPIResource(subaccount, storageboxsubaccount.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("storage_box_id"), testsupport.Int64ExactFromFunc(func() int64 { return storageBox.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return subaccount.ID })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("username"), testsupport.StringExactFromFunc(func() string { return subaccount.Username })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("server"), testsupport.StringExactFromFunc(func() string { return subaccount.Server })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("tf-e2e-subaccount")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("home_directory"), knownvalue.StringExact("updated")),
					statecheck.ExpectSensitiveValue(resOptional.TFID(), tfjsonpath.New("password")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("password"), knownvalue.StringExact(resOptional.Password)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("reachable_externally"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("samba_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("ssh_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("webdav_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("readonly"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
				},
			},
		},
	})
}
