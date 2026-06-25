package image_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/snapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func TestAccImageDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	testImageID, err := util.ParseID(teste2e.TestImageID)
	require.NoError(t, err)

	byName := &image.DData{
		Name: teste2e.TestImage,
	}
	byName.SetRName("by_name")
	byID := &image.DData{
		ID: teste2e.TestImageID,
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
		Name:              teste2e.TestImage,
		WithArchitecture:  "arm",
		IncludeDeprecated: true,
	}
	byName.SetRName("by_name")

	byLabel := &image.DData{
		WithSelector:      "!key",
		WithArchitecture:  "arm",
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

func TestAccImageDataSource_Snapshot(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcImage hcloud.Image

	srvs := server.NewBlueprint(t)

	snapshotRes := &snapshot.RData{
		ServerID:    srvs.ServerA.TFID() + ".id",
		Description: srvs.ServerA.Name,
		Labels: map[string]string{
			"name": randutil.GenerateID(),
		},
	}
	snapshotRes.SetRName("snapshot")

	byLabel := &image.DData{
		WithSelector: fmt.Sprintf("name=%s", snapshotRes.Labels["name"]),
		MostRecent:   new(true),

		Raw: fmt.Sprintf("depends_on = [%s]", snapshotRes.TFID()),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_snapshot", snapshotRes,
					"testdata/d/hcloud_image", byLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(snapshotRes.TFID(), snapshot.ByID(t, &hcImage)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("id"), testsupport.Int64ExactFromFunc(func() int64 { return hcImage.ID })),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("description"), knownvalue.StringExact(snapshotRes.Description)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"name": knownvalue.StringExact(snapshotRes.Labels["name"])})),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("snapshot")),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("architecture"), knownvalue.StringExact("x86")),
				},
			},
		},
	})
}

func TestAccImageDataSource_NotFound(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	byID := &image.DData{
		ID: "12345",
	}
	byID.SetRName("by_id")

	byName := &image.DData{
		Name: randutil.GenerateID(),
	}
	byName.SetRName("by_name")

	byLabel := &image.DData{
		WithSelector: fmt.Sprintf("name=%s", randutil.GenerateID()),
		MostRecent:   new(true),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byID,
				),
				ExpectError: regexp.MustCompile(`Resource \(image\) was not found: id=12345`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byName,
				),
				ExpectError: regexp.MustCompile(`Resource \(image\) was not found using name: [0-9a-f]{8}`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", byLabel,
				),
				ExpectError: regexp.MustCompile(`Resource \(image\) was not found using label selector: name=[0-9a-f]{8}

Query parameters: architecture=x86 label_selector=name=[0-9a-f]{8}
sort=created:desc`),
			},
		},
	})
}
