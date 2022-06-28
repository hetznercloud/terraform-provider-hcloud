package hcloud

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/datacenter"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/location"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/snapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_Resources(t *testing.T) {
	var provider = Provider()
	expectedResources := []string{
		certificate.ResourceType,
		firewall.ResourceType,
		firewall.AttachmentResourceType,
		certificate.UploadedResourceType,
		certificate.ManagedResourceType,
		floatingip.AssignmentResourceType,
		floatingip.ResourceType,
		primaryip.ResourceType,
		loadbalancer.NetworkResourceType,
		loadbalancer.ResourceType,
		loadbalancer.ServiceResourceType,
		loadbalancer.TargetResourceType,
		network.ResourceType,
		network.RouteResourceType,
		network.SubnetResourceType,
		rdns.ResourceType,
		server.NetworkResourceType,
		server.ResourceType,
		snapshot.ResourceType,
		sshkey.ResourceType,
		volume.AttachmentResourceType,
		volume.ResourceType,
		placementgroup.ResourceType,
	}

	resources := provider.Resources()
	assert.Len(t, resources, len(expectedResources))

	for _, datasource := range resources {
		assert.Contains(t, expectedResources, datasource.Name)
	}
}

func TestProvider_DataSources(t *testing.T) {
	var provider = Provider()
	expectedDataSources := []string{
		certificate.DataSourceType,
		certificate.DataSourceListType,
		datacenter.DataSourceType,
		datacenter.DataSourceListType,
		firewall.DataSourceType,
		firewall.DataSourceListType,
		floatingip.DataSourceType,
		floatingip.DataSourceListType,
		primaryip.DataSourceType,
		primaryip.DataSourceListType,
		image.DataSourceType,
		image.DataSourceListType,
		loadbalancer.DataSourceType,
		loadbalancer.DataSourceListType,
		location.DataSourceType,
		location.DataSourceListType,
		network.DataSourceType,
		network.DataSourceListType,
		placementgroup.DataSourceType,
		placementgroup.DataSourceListType,
		server.DataSourceType,
		server.DataSourceListType,
		servertype.DataSourceType,
		servertype.DataSourceListType,
		sshkey.DataSourceType,
		sshkey.DataSourceListType,
		volume.DataSourceType,
		volume.DataSourceListType,
	}

	dataSources := provider.DataSources()
	assert.Len(t, dataSources, len(expectedDataSources))

	for _, datasource := range dataSources {
		assert.Contains(t, expectedDataSources, datasource.Name)
	}
}
