package hcloud

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/datacenter"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/location"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storagebox"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsnapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxsubaccount"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxtype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/tflogutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zone"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zonerecord"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zonerrset"
)

type PluginProvider struct{}

var _ provider.Provider = &PluginProvider{}
var _ provider.ProviderWithActions = &PluginProvider{}
var _ provider.ProviderWithFunctions = &PluginProvider{}

func NewPluginProvider() provider.Provider {
	return &PluginProvider{}
}

// Metadata should return the metadata for the provider, such as
// a type name and version data.
//
// Implementing the MetadataResponse.TypeName will populate the
// datasource.MetadataRequest.ProviderTypeName and
// resource.MetadataRequest.ProviderTypeName fields automatically.
func (p *PluginProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hcloud"
	resp.Version = Version
}

// Schema should return the schema for this provider.
func (p *PluginProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "The Hetzner Cloud API token, can also be specified with the HCLOUD_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The Hetzner Cloud API endpoint, can be used to override the default API Endpoint https://api.hetzner.cloud/v1.",
				Optional:    true,
			},
			"endpoint_hetzner": schema.StringAttribute{
				Description: "The Hetzner API endpoint, can be used to override the default API Endpoint https://api.hetzner.com/v1.",
				Optional:    true,
			},
			"poll_interval": schema.StringAttribute{
				Description: "The interval at which actions are polled by the client. Default `500ms`. Increase this interval if you run into rate limiting errors.",
				Optional:    true,
			},
			"poll_function": schema.StringAttribute{
				Description: "The type of function to be used during the polling.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"constant", "exponential"}...),
				},
			},
		},
		// TODO: Uncomment once we get rid of the SDK v2 Provider
		// MarkdownDescription: `The Hetzner Cloud (hcloud) provider is used to interact with the resources supported by
		// [Hetzner Cloud](https://www.hetzner.com/cloud). The provider needs to be configured with the proper credentials
		//  before it can be used.`,
	}
}

// PluginProviderModel describes the provider data model.
type PluginProviderModel struct {
	Token           types.String `tfsdk:"token"`
	Endpoint        types.String `tfsdk:"endpoint"`
	EndpointHetzner types.String `tfsdk:"endpoint_hetzner"`
	PollInterval    types.String `tfsdk:"poll_interval"`
	PollFunction    types.String `tfsdk:"poll_function"`
}

// Configure is called at the beginning of the provider lifecycle, when
// Terraform sends to the provider the values the user specified in the
// provider configuration block. These are supplied in the
// ConfigureProviderRequest argument.
// Values from provider configuration are often used to initialize an
// API client, which should be stored on the struct to initialize an
// Provider interface.
func (p *PluginProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PluginProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := []hcloud.ClientOption{
		hcloud.WithApplication("hcloud-terraform", Version),
	}

	endpoint := os.Getenv("HCLOUD_ENDPOINT")
	if data.Endpoint.ValueString() != "" {
		endpoint = data.Endpoint.ValueString()
	}
	if endpoint != "" {
		opts = append(opts, hcloud.WithEndpoint(endpoint))
	}

	endpointHetzner := os.Getenv("HETZNER_ENDPOINT")
	if data.EndpointHetzner.ValueString() != "" {
		endpointHetzner = data.Endpoint.ValueString()
	}
	if endpointHetzner != "" {
		opts = append(opts, hcloud.WithHetznerEndpoint(endpointHetzner))
	}

	token := os.Getenv("HCLOUD_TOKEN")
	if data.Token.ValueString() != "" {
		token = data.Token.ValueString()
	}
	if token != "" {
		opts = append(opts, hcloud.WithToken(token))
	} else {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Hetzner Cloud API token",
			"While configuring the provider, the Hetzner Cloud API token was not found in the HCLOUD_TOKEN environment variable or provider configuration block token attribute.",
		)
	}

	pollOpts := hcloud.PollOpts{
		BackoffFunc: hcloud.ExponentialBackoffWithOpts(hcloud.ExponentialBackoffOpts{
			Base:       500 * time.Millisecond,
			Multiplier: 2.0,
			Cap:        10 * time.Second,
		}),
	}
	if data.PollInterval.ValueString() != "" {
		pollInterval, err := time.ParseDuration(data.PollInterval.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("poll_interval"),
				"Unparsable poll interval value",
				fmt.Sprintf("An unexpected error was encountered trying to parse the value.\n\n%s", err.Error()),
			)
		}
		if data.PollFunction.ValueString() == "constant" {
			pollOpts.BackoffFunc = hcloud.ConstantBackoff(pollInterval)
		} else {
			pollOpts.BackoffFunc = hcloud.ExponentialBackoff(2, pollInterval)
		}
	}
	opts = append(opts, hcloud.WithPollOpts(pollOpts))

	if resp.Diagnostics.HasError() {
		return
	}

	// Debug writer
	opts = append(opts,
		hcloud.WithDebugWriter(
			tflogutil.NewWriter(
				tflog.NewSubsystem(ctx, "hcloud-go", tflog.WithLevel(hclog.Debug)),
			),
		),
	)

	client := hcloud.NewClient(opts...)
	resp.DataSourceData = client
	resp.ResourceData = client
	resp.ActionData = client

	tflog.Info(ctx, "terraform-provider-hcloud info", map[string]any{"version": Version, "commit": Commit})
	tflog.Info(ctx, "hcloud-go info", map[string]any{"version": hcloud.Version})
}

// DataSources returns a slice of functions to instantiate each DataSource
// implementation.
//
// The data source type name is determined by the DataSource implementing
// the Metadata method. All data sources must have unique names.
func (p *PluginProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datacenter.NewDataSource,
		datacenter.NewDataSourceList,
		loadbalancertype.NewDataSource,
		loadbalancertype.NewDataSourceList,
		location.NewDataSource,
		location.NewDataSourceList,
		servertype.NewDataSource,
		servertype.NewDataSourceList,
		sshkey.NewDataSource,
		sshkey.NewDataSourceList,
		storagebox.NewDataSource,
		storagebox.NewDataSourceList,
		storageboxsnapshot.NewDataSource,
		storageboxsnapshot.NewDataSourceList,
		storageboxsubaccount.NewDataSource,
		storageboxsubaccount.NewDataSourceList,
		storageboxtype.NewDataSource,
		storageboxtype.NewDataSourceList,
		zone.NewDataSource,
		zone.NewDataSourceList,
		zonerrset.NewDataSource,
		zonerrset.NewDataSourceList,
	}
}

// Resources returns a slice of functions to instantiate each Resource
// implementation.
//
// The resource type name is determined by the Resource implementing
// the Metadata method. All resources must have unique names.
func (p *PluginProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		loadbalancer.NewNetworkResource,
		server.NewNetworkResource,
		sshkey.NewResource,
		storagebox.NewResource,
		storageboxsnapshot.NewResource,
		storageboxsubaccount.NewResource,
		zone.NewResource,
		zonerecord.NewResource,
		zonerrset.NewResource,
	}
}

func (p *PluginProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		server.NewPoweronAction,
		server.NewPoweroffAction,
		server.NewRebootAction,
		server.NewResetAction,
	}
}

func (p *PluginProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		zone.NewIDNAFunction,
		zonerrset.NewTXTRecordFunction,
	}
}
