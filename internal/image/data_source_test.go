package image_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func TestAccImageDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	testImageID, err := util.ParseID(teste2e.TestImageID)
	require.NoError(t, err)

	byName := &image.DData{
		ImageName: teste2e.TestImage,
	}
	byName.SetRName("by_name")
	byID := &image.DData{
		ImageID: teste2e.TestImageID,
	}
	byID.SetRName("by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byName,
					"testdata/d/hcloud_image", byID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("id"), knownvalue.Int64Exact(testImageID)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(teste2e.TestImage)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("system")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("created"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("system")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("Ubuntu 24.04")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("os_flavor"), knownvalue.StringExact("ubuntu")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("os_version"), knownvalue.StringExact("24.04")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("architecture"), knownvalue.StringExact("x86")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("rapid_deploy"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("labels"), knownvalue.MapSizeExact(0)),

					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("id"), knownvalue.Int64Exact(testImageID)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(teste2e.TestImage)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("system")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("created"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("system")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("Ubuntu 24.04")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("os_flavor"), knownvalue.StringExact("ubuntu")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("os_version"), knownvalue.StringExact("24.04")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("architecture"), knownvalue.StringExact("x86")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("rapid_deploy"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("labels"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccImageDataSource_WithFilters(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	byName := &image.DData{
		ImageName:         teste2e.TestImage,
		Architecture:      "arm",
		IncludeDeprecated: true,
	}
	byName.SetRName("by_name")

	byLabel := &image.DData{
		LabelSelector:     "!key",
		Architecture:      "arm",
		IncludeDeprecated: true,
		MostRecent:        new(true),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byName,
					"testdata/d/hcloud_image", byLabel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(teste2e.TestImage)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("system")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("architecture"), knownvalue.StringExact("arm")),

					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("system")),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("architecture"), knownvalue.StringExact("arm")),
				},
			},
		},
	})
}

func TestAccImageDataSource_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	byID := &image.DData{
		ImageID: teste2e.TestImageID,
	}
	byID.SetRName("by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.64.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byID,
					"testdata/r/terraform_data_resource", byID,
				),
			},
			{
				ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byID,
					"testdata/r/terraform_data_resource", byID,
				),
				PlanOnly: true,
			},
		},
	})
}
