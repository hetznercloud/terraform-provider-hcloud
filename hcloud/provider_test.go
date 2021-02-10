package hcloud

import (
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/datacenter"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/location"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/snapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
	"github.com/stretchr/testify/assert"
	"testing"
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
		floatingip.AssignmentResourceType,
		floatingip.ResourceType,
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
		datacenter.DatacentersDataSourceType,
		datacenter.DataSourceType,
		floatingip.DataSourceType,
		image.DataSourceType,
		loadbalancer.DataSourceType,
		location.DataSourceType,
		location.LocationsDataSourceType,
		network.DataSourceType,
		server.DataSourceType,
		servertype.DataSourceType,
		servertype.ServerTypesDataSourceType,
		sshkey.DataSourceType,
		sshkey.SSHKeysDataSourceType,
		volume.DataSourceType,
	}

	datasources := provider.DataSources()
	assert.Len(t, datasources, len(expectedDataSources))

	for _, datasource := range datasources {
		assert.Contains(t, expectedDataSources, datasource.Name)
	}
}
