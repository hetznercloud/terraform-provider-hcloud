package firewall

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAttachmentResource_Servers(t *testing.T) {
	var (
		srv hcloud.Server
		fw  hcloud.Firewall
	)

	fwRes := firewall.NewRData(t, "basic_firewall", nil, nil)
	srvRes := &server.RData{
		Name:  "test-server",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
	}
	srvRes.SetRName("test_server")

	fwAttRes := firewall.NewRDataAttachment("fw_ref", fwRes.TFID()+".id")
	fwAttRes.ServerIDRefs = append(fwAttRes.ServerIDRefs, srvRes.TFID()+".id")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  e2etests.PreCheck(t),
		Providers: e2etests.Providers(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
			testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, &fw)),
		),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvRes,
					"testdata/r/hcloud_firewall", fwRes,
					"testdata/r/hcloud_firewall_attachment", fwAttRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(srvRes.TFID(), server.ByID(t, &srv)),
					testsupport.CheckResourceExists(fwRes.TFID(), firewall.ByID(t, &fw)),
					testsupport.LiftTCF(hasServerResource(t, &fw, &srv)),
				),
			},
		},
	})
}

func TestAttachmentResource_LabelSelectors(t *testing.T) {
	var (
		srv hcloud.Server
		fw  hcloud.Firewall
	)

	fwRes := firewall.NewRData(t, "basic_firewall", nil, nil)
	srvRes := &server.RData{
		Name:  "test-server",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"firewall-attachment": "test-server",
		},
	}
	srvRes.SetRName("test_server")

	fwAttRes := firewall.NewRDataAttachment("fw_ref", fwRes.TFID()+".id")
	fwAttRes.LabelSelectors = append(fwAttRes.LabelSelectors, "firewall-attachment=test-server")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  e2etests.PreCheck(t),
		Providers: e2etests.Providers(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
			testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, &fw)),
		),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvRes,
					"testdata/r/hcloud_firewall", fwRes,
					"testdata/r/hcloud_firewall_attachment", fwAttRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(srvRes.TFID(), server.ByID(t, &srv)),
					testsupport.CheckResourceExists(fwRes.TFID(), firewall.ByID(t, &fw)),
					testsupport.LiftTCF(hasLabelSelectorResource(t, &fw, "firewall-attachment=test-server")),
				),
			},
		},
	})
}
