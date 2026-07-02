package pricing

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceType is the type name of the Hetzner Cloud Pricing data source.
const DataSourceType = "hcloud_pricing"

var _ datasource.DataSource = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

type DataSource struct {
	client *hcloud.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata should return the full name of the data source.
func (d *DataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// priceSchema returns the schema for a net/gross price including the currency
// and VAT rate.
func priceSchema(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: description,
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"currency": schema.StringAttribute{
				MarkdownDescription: "Currency the price is denominated in.",
				Computed:            true,
			},
			"vat_rate": schema.StringAttribute{
				MarkdownDescription: "VAT rate in percent used to calculate the gross price from the net price.",
				Computed:            true,
			},
			"net": schema.StringAttribute{
				MarkdownDescription: "Price without VAT.",
				Computed:            true,
			},
			"gross": schema.StringAttribute{
				MarkdownDescription: "Price with VAT added.",
				Computed:            true,
			},
		},
	}
}

// primaryIPPriceSchema returns the schema for a net/gross Primary IP price.
func primaryIPPriceSchema(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: description,
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"net": schema.StringAttribute{
				MarkdownDescription: "Price without VAT.",
				Computed:            true,
			},
			"gross": schema.StringAttribute{
				MarkdownDescription: "Price with VAT added.",
				Computed:            true,
			},
		},
	}
}

// perGBPricingSchema returns the schema for a per GB and month price.
func perGBPricingSchema(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: description,
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"per_gb_month": priceSchema("Price per GB and month."),
		},
	}
}

// Schema should return the schema for this data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides prices for the resources offered by the Hetzner Cloud.

The prices are returned for the Hetzner Cloud project the API token used to configure the provider belongs to.

See the [Pricing API documentation](https://docs.hetzner.cloud/reference/cloud#pricing) for more details.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"location": schema.StringAttribute{
			MarkdownDescription: "Only return per-location prices for the location with this name (e.g. `fsn1`).",
			Optional:            true,
		},
		"server_type": schema.StringAttribute{
			MarkdownDescription: "Only return prices for the Server Type with this name (e.g. `cx22`).",
			Optional:            true,
		},
		"load_balancer_type": schema.StringAttribute{
			MarkdownDescription: "Only return prices for the Load Balancer Type with this name (e.g. `lb11`).",
			Optional:            true,
		},
		"currency": schema.StringAttribute{
			MarkdownDescription: "Currency all prices are denominated in.",
			Computed:            true,
		},
		"vat_rate": schema.StringAttribute{
			MarkdownDescription: "VAT rate in percent used to calculate the gross prices from the net prices.",
			Computed:            true,
		},
		"image": perGBPricingSchema("Price of the disk space used by Images (Snapshots and Backups)."),
		"floating_ips": schema.ListNestedAttribute{
			MarkdownDescription: "Prices of the Floating IPs, grouped by type.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of the Floating IP (`ipv4` or `ipv6`).",
						Computed:            true,
					},
					"prices": schema.ListNestedAttribute{
						MarkdownDescription: "Prices of the Floating IP type, per location.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"location": schema.StringAttribute{
									MarkdownDescription: "Name of the location the price applies to.",
									Computed:            true,
								},
								"monthly": priceSchema("Monthly costs for a Floating IP type in this location."),
							},
						},
					},
				},
			},
		},
		"primary_ips": schema.ListNestedAttribute{
			MarkdownDescription: "Prices of the Primary IPs, grouped by type.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of the Primary IP (`ipv4` or `ipv6`).",
						Computed:            true,
					},
					"prices": schema.ListNestedAttribute{
						MarkdownDescription: "Prices of the Primary IP type, per location.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"location": schema.StringAttribute{
									MarkdownDescription: "Name of the location the price applies to.",
									Computed:            true,
								},
								"hourly":  primaryIPPriceSchema("Hourly costs for a Primary IP type in this location."),
								"monthly": primaryIPPriceSchema("Monthly costs for a Primary IP type in this location."),
							},
						},
					},
				},
			},
		},
		"server_backup": schema.SingleNestedAttribute{
			MarkdownDescription: "Will increase base server costs by specific percentage if server backups are enabled.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"percentage": schema.StringAttribute{
					MarkdownDescription: "Percentage by which the base server costs increase if server backups are enabled.",
					Computed:            true,
				},
			},
		},
		"server_types":        typePricingSchema("Server Type"),
		"load_balancer_types": typePricingSchema("Load Balancer Type"),
		"volume":              perGBPricingSchema("Price of the disk space used by Volumes."),
	}
}

// typePricingSchema returns the schema for the per-location pricing of a Server
// Type or Load Balancer Type.
func typePricingSchema(name string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Prices of the " + name + "s, grouped by type.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.Int64Attribute{
					MarkdownDescription: "ID of the " + name + ".",
					Computed:            true,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the " + name + ".",
					Computed:            true,
				},
				"prices": schema.ListNestedAttribute{
					MarkdownDescription: "Prices of the " + name + ", per location.",
					Computed:            true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"location": schema.StringAttribute{
								MarkdownDescription: "Name of the location the price applies to.",
								Computed:            true,
							},
							"hourly":  priceSchema("Hourly costs for a " + name + " in this location."),
							"monthly": priceSchema("Monthly costs for a " + name + " in this location."),
							"included_traffic": schema.Int64Attribute{
								MarkdownDescription: "Free traffic per month in bytes.",
								Computed:            true,
							},
							"per_tb_traffic": priceSchema("Costs per additional TB of traffic in this location."),
						},
					},
				},
			},
		},
	}
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, _, err := d.client.Pricing.Get(ctx)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, filterPricing(result, data))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
