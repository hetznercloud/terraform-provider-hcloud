package pricing

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

// model is the top level model for the Hetzner Cloud pricing data source.
type model struct {
	// Filters
	Location         types.String `tfsdk:"location"`
	ServerType       types.String `tfsdk:"server_type"`
	LoadBalancerType types.String `tfsdk:"load_balancer_type"`

	Currency          types.String `tfsdk:"currency"`
	VATRate           types.String `tfsdk:"vat_rate"`
	Image             types.Object `tfsdk:"image"`
	FloatingIPs       types.List   `tfsdk:"floating_ips"`
	PrimaryIPs        types.List   `tfsdk:"primary_ips"`
	ServerBackup      types.Object `tfsdk:"server_backup"`
	ServerTypes       types.List   `tfsdk:"server_types"`
	LoadBalancerTypes types.List   `tfsdk:"load_balancer_types"`
	Volume            types.Object `tfsdk:"volume"`
}

var _ util.ModelFromAPI[hcloud.Pricing] = &model{}

var (
	_ util.ModelFromAPI[hcloud.Price]                         = (*priceModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*priceModel)(nil)
	_ util.ModelFromAPI[hcloud.PrimaryIPPrice]                = (*primaryIPPriceModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*primaryIPPriceModel)(nil)
	_ util.ModelFromAPI[hcloud.Price]                         = (*perGBPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*perGBPricingModel)(nil)
	_ util.ModelFromAPI[hcloud.ServerBackupPricing]           = (*serverBackupPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*serverBackupPricingModel)(nil)
	_ util.ModelFromAPI[hcloud.FloatingIPTypePricing]         = (*floatingIPPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*floatingIPPricingModel)(nil)
	_ util.ModelFromAPI[hcloud.FloatingIPTypeLocationPricing] = (*floatingIPLocationPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*floatingIPLocationPricingModel)(nil)
	_ util.ModelFromAPI[hcloud.PrimaryIPPricing]              = (*primaryIPPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*primaryIPPricingModel)(nil)
	_ util.ModelFromAPI[hcloud.PrimaryIPTypePricing]          = (*primaryIPLocationPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*primaryIPLocationPricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*typePricingModel)(nil)
	_ util.ModelToTerraform[types.Object]                     = (*typeLocationPricingModel)(nil)
)

func (m *model) FromAPI(ctx context.Context, hc hcloud.Pricing) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Currency = types.StringValue(hc.Currency)
	m.VATRate = types.StringValue(hc.VATRate)

	{
		value := perGBPricingModel{}
		diags.Append(value.FromAPI(ctx, hc.Image.PerGBMonth)...)

		m.Image, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	{
		value := perGBPricingModel{}
		diags.Append(value.FromAPI(ctx, hc.Volume.PerGBMonthly)...)

		m.Volume, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	{
		value := serverBackupPricingModel{}
		diags.Append(value.FromAPI(ctx, hc.ServerBackup)...)

		m.ServerBackup, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	m.FloatingIPs, newDiags = listValueFromAPI(ctx, (&floatingIPPricingModel{}).tfType(), hc.FloatingIPs, floatingIPPricingObjectFromAPI)
	diags.Append(newDiags...)

	m.PrimaryIPs, newDiags = listValueFromAPI(ctx, (&primaryIPPricingModel{}).tfType(), hc.PrimaryIPs, primaryIPPricingObjectFromAPI)
	diags.Append(newDiags...)

	m.ServerTypes, newDiags = listValueFromAPI(ctx, (&typePricingModel{}).tfType(), hc.ServerTypes, serverTypePricingObjectFromAPI)
	diags.Append(newDiags...)

	m.LoadBalancerTypes, newDiags = listValueFromAPI(ctx, (&typePricingModel{}).tfType(), hc.LoadBalancerTypes, loadBalancerTypePricingObjectFromAPI)
	diags.Append(newDiags...)

	return diags
}

// priceModel represents a net/gross price including the currency and VAT rate.
type priceModel struct {
	Currency types.String `tfsdk:"currency"`
	VATRate  types.String `tfsdk:"vat_rate"`
	Net      types.String `tfsdk:"net"`
	Gross    types.String `tfsdk:"gross"`
}

func (m *priceModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"currency": types.StringType,
		"vat_rate": types.StringType,
		"net":      types.StringType,
		"gross":    types.StringType,
	}
}

func (m *priceModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *priceModel) FromAPI(_ context.Context, hc hcloud.Price) diag.Diagnostics {
	m.Currency = types.StringValue(hc.Currency)
	m.VATRate = types.StringValue(hc.VATRate)
	m.Net = types.StringValue(hc.Net)
	m.Gross = types.StringValue(hc.Gross)

	return nil
}

func (m *priceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

// priceObjectFromAPI converts a single hcloud.Price into its Terraform object.
func priceObjectFromAPI(ctx context.Context, hc hcloud.Price) (types.Object, diag.Diagnostics) {
	value := priceModel{}
	diags := value.FromAPI(ctx, hc)

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// primaryIPPriceModel represents a net/gross price for Primary IPs. Unlike
// priceModel it does not include the currency or VAT rate.
type primaryIPPriceModel struct {
	Net   types.String `tfsdk:"net"`
	Gross types.String `tfsdk:"gross"`
}

func (m *primaryIPPriceModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"net":   types.StringType,
		"gross": types.StringType,
	}
}

func (m *primaryIPPriceModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *primaryIPPriceModel) FromAPI(_ context.Context, hc hcloud.PrimaryIPPrice) diag.Diagnostics {
	m.Net = types.StringValue(hc.Net)
	m.Gross = types.StringValue(hc.Gross)

	return nil
}

func (m *primaryIPPriceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func primaryIPPriceObjectFromAPI(ctx context.Context, hc hcloud.PrimaryIPPrice) (types.Object, diag.Diagnostics) {
	value := primaryIPPriceModel{}
	diags := value.FromAPI(ctx, hc)

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// perGBPricingModel represents a per GB and month price, used by Images and
// Volumes.
type perGBPricingModel struct {
	PerGBMonth types.Object `tfsdk:"per_gb_month"`
}

func (m *perGBPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"per_gb_month": (&priceModel{}).tfType(),
	}
}

func (m *perGBPricingModel) FromAPI(ctx context.Context, hc hcloud.Price) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.PerGBMonth, newDiags = priceObjectFromAPI(ctx, hc)
	diags.Append(newDiags...)

	return diags
}

func (m *perGBPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

// serverBackupPricingModel represents the surcharge for server backups.
type serverBackupPricingModel struct {
	Percentage types.String `tfsdk:"percentage"`
}

func (m *serverBackupPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"percentage": types.StringType,
	}
}

func (m *serverBackupPricingModel) FromAPI(_ context.Context, hc hcloud.ServerBackupPricing) diag.Diagnostics {
	m.Percentage = types.StringValue(hc.Percentage)

	return nil
}

func (m *serverBackupPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

// floatingIPPricingModel represents the pricing of a Floating IP type across
// all locations.
type floatingIPPricingModel struct {
	Type   types.String `tfsdk:"type"`
	Prices types.List   `tfsdk:"prices"`
}

func (m *floatingIPPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":   types.StringType,
		"prices": types.ListType{ElemType: (&floatingIPLocationPricingModel{}).tfType()},
	}
}

func (m *floatingIPPricingModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *floatingIPPricingModel) FromAPI(ctx context.Context, hc hcloud.FloatingIPTypePricing) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Type = types.StringValue(string(hc.Type))

	m.Prices, newDiags = listValueFromAPI(ctx, (&floatingIPLocationPricingModel{}).tfType(), hc.Pricings, floatingIPLocationPricingObjectFromAPI)
	diags.Append(newDiags...)

	return diags
}

func (m *floatingIPPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func floatingIPPricingObjectFromAPI(ctx context.Context, hc hcloud.FloatingIPTypePricing) (types.Object, diag.Diagnostics) {
	value := floatingIPPricingModel{}
	diags := value.FromAPI(ctx, hc)

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// floatingIPLocationPricingModel represents the pricing of a Floating IP type
// at a single location.
type floatingIPLocationPricingModel struct {
	Location types.String `tfsdk:"location"`
	Monthly  types.Object `tfsdk:"monthly"`
}

func (m *floatingIPLocationPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"location": types.StringType,
		"monthly":  (&priceModel{}).tfType(),
	}
}

func (m *floatingIPLocationPricingModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *floatingIPLocationPricingModel) FromAPI(ctx context.Context, hc hcloud.FloatingIPTypeLocationPricing) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Location = types.StringValue(locationName(hc.Location))

	m.Monthly, newDiags = priceObjectFromAPI(ctx, hc.Monthly)
	diags.Append(newDiags...)

	return diags
}

func (m *floatingIPLocationPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func floatingIPLocationPricingObjectFromAPI(ctx context.Context, hc hcloud.FloatingIPTypeLocationPricing) (types.Object, diag.Diagnostics) {
	value := floatingIPLocationPricingModel{}
	diags := value.FromAPI(ctx, hc)

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// primaryIPPricingModel represents the pricing of a Primary IP type across all
// locations.
type primaryIPPricingModel struct {
	Type   types.String `tfsdk:"type"`
	Prices types.List   `tfsdk:"prices"`
}

func (m *primaryIPPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":   types.StringType,
		"prices": types.ListType{ElemType: (&primaryIPLocationPricingModel{}).tfType()},
	}
}

func (m *primaryIPPricingModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *primaryIPPricingModel) FromAPI(ctx context.Context, hc hcloud.PrimaryIPPricing) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Type = types.StringValue(hc.Type)

	m.Prices, newDiags = listValueFromAPI(ctx, (&primaryIPLocationPricingModel{}).tfType(), hc.Pricings, primaryIPLocationPricingObjectFromAPI)
	diags.Append(newDiags...)

	return diags
}

func (m *primaryIPPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func primaryIPPricingObjectFromAPI(ctx context.Context, hc hcloud.PrimaryIPPricing) (types.Object, diag.Diagnostics) {
	value := primaryIPPricingModel{}
	diags := value.FromAPI(ctx, hc)

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// primaryIPLocationPricingModel represents the pricing of a Primary IP type at
// a single location.
type primaryIPLocationPricingModel struct {
	Location types.String `tfsdk:"location"`
	Hourly   types.Object `tfsdk:"hourly"`
	Monthly  types.Object `tfsdk:"monthly"`
}

func (m *primaryIPLocationPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"location": types.StringType,
		"hourly":   (&primaryIPPriceModel{}).tfType(),
		"monthly":  (&primaryIPPriceModel{}).tfType(),
	}
}

func (m *primaryIPLocationPricingModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *primaryIPLocationPricingModel) FromAPI(ctx context.Context, hc hcloud.PrimaryIPTypePricing) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Location = types.StringValue(hc.Location)

	m.Hourly, newDiags = primaryIPPriceObjectFromAPI(ctx, hc.Hourly)
	diags.Append(newDiags...)

	m.Monthly, newDiags = primaryIPPriceObjectFromAPI(ctx, hc.Monthly)
	diags.Append(newDiags...)

	return diags
}

func (m *primaryIPLocationPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func primaryIPLocationPricingObjectFromAPI(ctx context.Context, hc hcloud.PrimaryIPTypePricing) (types.Object, diag.Diagnostics) {
	value := primaryIPLocationPricingModel{}
	diags := value.FromAPI(ctx, hc)

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// typePricingModel represents the pricing of a Server Type or Load Balancer
// Type across all locations.
type typePricingModel struct {
	ID     types.Int64  `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Prices types.List   `tfsdk:"prices"`
}

func (m *typePricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.Int64Type,
		"name":   types.StringType,
		"prices": types.ListType{ElemType: (&typeLocationPricingModel{}).tfType()},
	}
}

func (m *typePricingModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *typePricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func serverTypePricingObjectFromAPI(ctx context.Context, hc hcloud.ServerTypePricing) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := typePricingModel{}
	if hc.ServerType != nil {
		value.ID = types.Int64Value(hc.ServerType.ID)
		value.Name = types.StringValue(hc.ServerType.Name)
	}

	prices, newDiags := listValueFromAPI(ctx, (&typeLocationPricingModel{}).tfType(), hc.Pricings,
		func(ctx context.Context, p hcloud.ServerTypeLocationPricing) (types.Object, diag.Diagnostics) {
			item := typeLocationPricingModel{}
			diags := item.FromAPI(ctx, locationName(p.Location), p.Hourly, p.Monthly, p.IncludedTraffic, p.PerTBTraffic)

			obj, newDiags := item.ToTerraform(ctx)
			diags.Append(newDiags...)

			return obj, diags
		},
	)
	diags.Append(newDiags...)
	value.Prices = prices

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

func loadBalancerTypePricingObjectFromAPI(ctx context.Context, hc hcloud.LoadBalancerTypePricing) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := typePricingModel{}
	if hc.LoadBalancerType != nil {
		value.ID = types.Int64Value(hc.LoadBalancerType.ID)
		value.Name = types.StringValue(hc.LoadBalancerType.Name)
	}

	prices, newDiags := listValueFromAPI(ctx, (&typeLocationPricingModel{}).tfType(), hc.Pricings,
		func(ctx context.Context, p hcloud.LoadBalancerTypeLocationPricing) (types.Object, diag.Diagnostics) {
			item := typeLocationPricingModel{}
			diags := item.FromAPI(ctx, locationName(p.Location), p.Hourly, p.Monthly, p.IncludedTraffic, p.PerTBTraffic)

			obj, newDiags := item.ToTerraform(ctx)
			diags.Append(newDiags...)

			return obj, diags
		},
	)
	diags.Append(newDiags...)
	value.Prices = prices

	obj, newDiags := value.ToTerraform(ctx)
	diags.Append(newDiags...)

	return obj, diags
}

// typeLocationPricingModel represents the pricing of a Server Type or Load
// Balancer Type at a single location.
type typeLocationPricingModel struct {
	Location        types.String `tfsdk:"location"`
	Hourly          types.Object `tfsdk:"hourly"`
	Monthly         types.Object `tfsdk:"monthly"`
	IncludedTraffic types.Int64  `tfsdk:"included_traffic"`
	PerTBTraffic    types.Object `tfsdk:"per_tb_traffic"`
}

func (m *typeLocationPricingModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"location":         types.StringType,
		"hourly":           (&priceModel{}).tfType(),
		"monthly":          (&priceModel{}).tfType(),
		"included_traffic": types.Int64Type,
		"per_tb_traffic":   (&priceModel{}).tfType(),
	}
}

func (m *typeLocationPricingModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *typeLocationPricingModel) FromAPI(ctx context.Context, location string, hourly, monthly hcloud.Price, includedTraffic uint64, perTBTraffic hcloud.Price) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Location = types.StringValue(location)

	m.Hourly, newDiags = priceObjectFromAPI(ctx, hourly)
	diags.Append(newDiags...)

	m.Monthly, newDiags = priceObjectFromAPI(ctx, monthly)
	diags.Append(newDiags...)

	m.IncludedTraffic = types.Int64Value(int64(includedTraffic)) //nolint:gosec // traffic in bytes always fits in int64

	m.PerTBTraffic, newDiags = priceObjectFromAPI(ctx, perTBTraffic)
	diags.Append(newDiags...)

	return diags
}

func (m *typeLocationPricingModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

// listValueFromAPI converts a slice of API values into a Terraform list of
// objects, using the provided conversion function for each element.
func listValueFromAPI[API any](
	ctx context.Context,
	elemType attr.Type,
	items []API,
	convert func(context.Context, API) (types.Object, diag.Diagnostics),
) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		obj, newDiags := convert(ctx, item)
		diags.Append(newDiags...)

		values = append(values, obj)
	}

	list, newDiags := types.ListValue(elemType, values)
	diags.Append(newDiags...)

	return list, diags
}

// locationName returns the name of the given location, or an empty string if it
// is not set.
func locationName(location *hcloud.Location) string {
	if location == nil {
		return ""
	}
	return location.Name
}

// filterPricing applies the optional filters configured on the data source. The
// Pricing API does not support server side filtering, so the full response is
// fetched and filtered locally.
func filterPricing(p hcloud.Pricing, data model) hcloud.Pricing {
	location := data.Location.ValueString()
	serverType := data.ServerType.ValueString()
	loadBalancerType := data.LoadBalancerType.ValueString()

	if location != "" {
		for i := range p.FloatingIPs {
			p.FloatingIPs[i].Pricings = filterSlice(p.FloatingIPs[i].Pricings, func(v hcloud.FloatingIPTypeLocationPricing) bool {
				return locationName(v.Location) == location
			})
		}
		for i := range p.PrimaryIPs {
			p.PrimaryIPs[i].Pricings = filterSlice(p.PrimaryIPs[i].Pricings, func(v hcloud.PrimaryIPTypePricing) bool {
				return v.Location == location
			})
		}
		for i := range p.ServerTypes {
			p.ServerTypes[i].Pricings = filterSlice(p.ServerTypes[i].Pricings, func(v hcloud.ServerTypeLocationPricing) bool {
				return locationName(v.Location) == location
			})
		}
		for i := range p.LoadBalancerTypes {
			p.LoadBalancerTypes[i].Pricings = filterSlice(p.LoadBalancerTypes[i].Pricings, func(v hcloud.LoadBalancerTypeLocationPricing) bool {
				return locationName(v.Location) == location
			})
		}
	}

	if serverType != "" {
		p.ServerTypes = filterSlice(p.ServerTypes, func(v hcloud.ServerTypePricing) bool {
			return v.ServerType != nil && v.ServerType.Name == serverType
		})
	}

	if loadBalancerType != "" {
		p.LoadBalancerTypes = filterSlice(p.LoadBalancerTypes, func(v hcloud.LoadBalancerTypePricing) bool {
			return v.LoadBalancerType != nil && v.LoadBalancerType.Name == loadBalancerType
		})
	}

	return p
}

// filterSlice returns a new slice containing only the items for which keep
// returns true.
func filterSlice[T any](items []T, keep func(T) bool) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			result = append(result, item)
		}
	}
	return result
}
