package storagebox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storagebox"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	storageBox := &hcloud.StorageBox{}

	res := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           fmt.Sprintf("snapshot-%s", randutil.GenerateID()),
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxType},
			Location:       schema.Location{Name: teste2e.TestLocationName},
			Labels: map[string]string{
				"key": "value",
			},
		},
		Password: storagebox.GeneratePassword(t),
	}
	res.SetRName("default")

	resOptional := &storagebox.RData{
		StorageBox: schema.StorageBox{
			Name:           res.Name + "-updated",
			StorageBoxType: schema.StorageBoxType{Name: teste2e.TestStorageBoxTypeUpgrade},
			Location:       res.Location,
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Password: storagebox.GeneratePassword(t), // Also test password update
		Raw: `
			access_settings = {
				reachable_externally = true
				samba_enabled        = true
				ssh_enabled          = true
				webdav_enabled       = true
				zfs_enabled          = true
			}

			delete_protection = true

			snapshot_plan = {
				max_snapshots = 10
				minute        = 16
				hour          = 18
				day_of_week   = 3
			}`,
	}
	resOptional.SetRName(res.RName())

	sshKey := sshkey.NewRData(t, "storage-box")

	resWithSSHKey := &storagebox.RData{
		StorageBox: resOptional.StorageBox,
		Password:   resOptional.Password,
		Raw:        resOptional.Raw,

		SSHKeys: []string{sshKey.PublicKey},
	}
	resWithSSHKey.SetRName(resOptional.RName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(storagebox.ResourceType, storagebox.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create with only required attributes

				Config: tmplMan.Render(t, "testdata/r/hcloud_storage_box", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(res.Name)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("username"), testsupport.StringExactFromFunc(func() string { return storageBox.Username })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxType)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectSensitiveValue(res.TFID(), tfjsonpath.New("password")),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("password"), knownvalue.StringExact(res.Password)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("ssh_keys"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("reachable_externally"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("samba_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("ssh_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("webdav_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("access_settings").AtMapKey("zfs_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("server"), testsupport.StringExactFromFunc(func() string { return storageBox.Server })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("system"), testsupport.StringExactFromFunc(func() string { return storageBox.System })),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res.TFID(), tfjsonpath.New("snapshot_plan"), knownvalue.Null())},
			},
			{
				// Import

				Config:                  tmplMan.Render(t, "testdata/r/hcloud_storage_box", resOptional),
				ImportState:             true,
				ResourceName:            res.TFID(),
				ImportStateId:           res.Name,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // Not returned in the API
			},
			{
				// Update with all optional attributes

				Config: tmplMan.Render(t, "testdata/r/hcloud_storage_box", resOptional),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(resOptional.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resOptional.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(resOptional.Name)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("username"), testsupport.StringExactFromFunc(func() string { return storageBox.Username })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxTypeUpgrade)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectSensitiveValue(resOptional.TFID(), tfjsonpath.New("password")),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("password"), knownvalue.StringExact(resOptional.Password)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"foo": knownvalue.StringExact("bar")})),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("ssh_keys"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("reachable_externally"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("samba_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("ssh_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("webdav_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("access_settings").AtMapKey("zfs_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("server"), testsupport.StringExactFromFunc(func() string { return storageBox.Server })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("system"), testsupport.StringExactFromFunc(func() string { return storageBox.System })),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("max_snapshots"), knownvalue.Int32Exact(10)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("minute"), knownvalue.Int32Exact(16)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("hour"), knownvalue.Int32Exact(18)),
					statecheck.ExpectKnownValue(resOptional.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("day_of_week"), knownvalue.Int32Exact(3)),
				},
			},
			{
				// Validate changing SSH Key attribute is not applied
				Config: tmplMan.Render(t, "testdata/r/hcloud_storage_box", resWithSSHKey),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure it's actually doing nothing
						plancheck.ExpectResourceAction(resOptional.TFID(), plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// And check that the state still has an empty set for the ssh_keys
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("ssh_keys"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				// Create with all optional attributes
				Taint:  []string{resWithSSHKey.TFID()}, // replace the resource
				Config: tmplMan.Render(t, "testdata/r/hcloud_storage_box", resWithSSHKey),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resWithSSHKey.TFID(), testsupport.CopyAPIResource(storageBox, storagebox.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(resWithSSHKey.Name)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("username"), testsupport.StringExactFromFunc(func() string { return storageBox.Username })),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("storage_box_type"), knownvalue.StringExact(teste2e.TestStorageBoxTypeUpgrade)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectSensitiveValue(resWithSSHKey.TFID(), tfjsonpath.New("password")),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("password"), knownvalue.StringExact(resWithSSHKey.Password)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"foo": knownvalue.StringExact("bar")})),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("ssh_keys"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("ssh_keys").AtSliceIndex(0), knownvalue.StringExact(sshKey.PublicKey)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("access_settings").AtMapKey("reachable_externally"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("access_settings").AtMapKey("samba_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("access_settings").AtMapKey("ssh_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("access_settings").AtMapKey("webdav_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("access_settings").AtMapKey("zfs_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("server"), testsupport.StringExactFromFunc(func() string { return storageBox.Server })),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("system"), testsupport.StringExactFromFunc(func() string { return storageBox.System })),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("max_snapshots"), knownvalue.Int32Exact(10)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("minute"), knownvalue.Int32Exact(16)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("hour"), knownvalue.Int32Exact(18)),
					statecheck.ExpectKnownValue(resWithSSHKey.TFID(), tfjsonpath.New("snapshot_plan").AtMapKey("day_of_week"), knownvalue.Int32Exact(3)),
				},
			},
		},
	})
}
