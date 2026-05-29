package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

type networkResourceData struct {
	ID             types.String      `tfsdk:"id"`
	LoadBalancerID types.Int64       `tfsdk:"load_balancer_id"`
	NetworkID      types.Int64       `tfsdk:"network_id"`
	SubnetID       types.String      `tfsdk:"subnet_id"`
	IP             iptypes.IPAddress `tfsdk:"ip"`

	EnablePublicInterface types.Bool `tfsdk:"enable_public_interface"`
}

// nolint:unparam
func populateNetworkResourceData(
	_ context.Context,
	data *networkResourceData,
	loadBalancer *hcloud.LoadBalancer,
	attachment *hcloud.LoadBalancerPrivateNet,
) diag.Diagnostics {
	data.ID = types.StringValue(fmt.Sprintf("%d-%d", loadBalancer.ID, attachment.Network.ID))
	data.LoadBalancerID = types.Int64Value(loadBalancer.ID)
	data.NetworkID = types.Int64Value(attachment.Network.ID)
	data.IP = iptypes.NewIPAddressValue(attachment.IP.String())

	data.EnablePublicInterface = types.BoolValue(loadBalancer.PublicNet.Enabled)

	return nil
}

type serviceModel struct {
	ID              types.String `tfsdk:"id"`
	LoadBalancerID  types.Int64  `tfsdk:"load_balancer_id"`
	Protocol        types.String `tfsdk:"protocol"`
	ListenPort      types.Int32  `tfsdk:"listen_port"`
	DestinationPort types.Int32  `tfsdk:"destination_port"`
	Proxyprotocol   types.Bool   `tfsdk:"proxyprotocol"`
	HTTP            types.Object `tfsdk:"http"`
	HealthCheck     types.Object `tfsdk:"health_check"`
}

func (m *serviceModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"load_balancer_id": types.Int64Type,
		"protocol":         types.StringType,
		"listen_port":      types.Int32Type,
		"destination_port": types.Int32Type,
		"proxyprotocol":    types.BoolType,
		"http":             (&serviceModelHTTP{}).tfType(),
		"health_check":     (&serviceModelHealthCheck{}).tfType(),
	}
}

func (m *serviceModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.LoadBalancerService] = &serviceModel{}
var _ util.ModelToTerraform[types.Object] = &serviceModel{}

func (m *serviceModel) FromAPI(ctx context.Context, hc *hcloud.LoadBalancerService) diag.Diagnostics {
	var diags, newDiags diag.Diagnostics

	m.ListenPort = types.Int32Value(int32(hc.ListenPort))           //nolint:gosec
	m.DestinationPort = types.Int32Value(int32(hc.DestinationPort)) //nolint:gosec
	m.Proxyprotocol = types.BoolValue(hc.Proxyprotocol)
	m.Protocol = types.StringValue(string(hc.Protocol))

	{
		value := serviceModelHTTP{}
		diags.Append(value.FromAPI(ctx, &hc.HTTP)...)

		m.HTTP, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	{
		value := serviceModelHealthCheck{}
		diags.Append(value.FromAPI(ctx, &hc.HealthCheck)...)

		m.HealthCheck, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	return nil
}

func (m *serviceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type serviceModelHTTP struct {
	StickySessions types.Bool   `tfsdk:"sticky_sessions"`
	CookieName     types.String `tfsdk:"cookie_name"`
	CookieLifetime types.Int32  `tfsdk:"cookie_lifetime"`
	CertificateIDs types.List   `tfsdk:"certificate_ids"`
	RedirectHTTP   types.Bool   `tfsdk:"redirect_http"`
	TimeoutIdle    types.Int32  `tfsdk:"timeout_idle"`
}

func (m *serviceModelHTTP) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sticky_sessions": types.BoolType,
		"cookie_name":     types.StringType,
		"cookie_lifetime": types.Int32Type,
		"certificate_ids": types.ListType{ElemType: types.Int64Type},
		"redirect_http":   types.BoolType,
		"timeout_idle":    types.Int32Type,
	}
}

func (m *serviceModelHTTP) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.LoadBalancerServiceHTTP] = &serviceModelHTTP{}
var _ util.ModelToTerraform[types.Object] = &serviceModelHTTP{}

func (m *serviceModelHTTP) FromAPI(_ context.Context, hc *hcloud.LoadBalancerServiceHTTP) diag.Diagnostics {
	var diags, newDiags diag.Diagnostics

	var httpCertIDs []attr.Value
	for _, cert := range hc.Certificates {
		httpCertIDs = append(httpCertIDs, types.Int64Value(cert.ID))
	}

	m.StickySessions = types.BoolValue(hc.StickySessions)
	m.CookieName = types.StringValue(hc.CookieName)
	m.CookieLifetime = types.Int32Value(int32(hc.CookieLifetime.Seconds()))
	m.CertificateIDs, newDiags = types.ListValue(types.Int64Type, httpCertIDs)
	diags = append(diags, newDiags...)
	m.RedirectHTTP = types.BoolValue(hc.RedirectHTTP)
	m.TimeoutIdle = types.Int32Value(int32(hc.TimeoutIdle.Seconds()))

	return diags
}

func (m *serviceModelHTTP) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type serviceModelHealthCheck struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int32  `tfsdk:"port"`
	Interval types.Int32  `tfsdk:"interval"`
	Timeout  types.Int32  `tfsdk:"timeout"`
	Retries  types.Int32  `tfsdk:"retries"`
	HTTP     types.Object `tfsdk:"http"`
}

func (m *serviceModelHealthCheck) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"protocol": types.StringType,
		"port":     types.Int32Type,
		"interval": types.Int32Type,
		"timeout":  types.Int32Type,
		"retries":  types.Int32Type,
		"http":     (&serviceModelHealthCheckHTTP{}).tfType(),
	}
}

func (m *serviceModelHealthCheck) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.LoadBalancerServiceHealthCheck] = &serviceModelHealthCheck{}
var _ util.ModelToTerraform[types.Object] = &serviceModelHealthCheck{}

func (m *serviceModelHealthCheck) FromAPI(ctx context.Context, hc *hcloud.LoadBalancerServiceHealthCheck) diag.Diagnostics {
	var diags, newDiags diag.Diagnostics

	m.Protocol = types.StringValue(string(hc.Protocol))
	m.Port = types.Int32Value(int32(hc.Port)) //nolint:gosec
	m.Interval = types.Int32Value(int32(hc.Interval.Seconds()))
	m.Timeout = types.Int32Value(int32(hc.Timeout.Seconds()))
	m.Retries = types.Int32Value(int32(hc.Retries)) //nolint:gosec

	if hc.HTTP != nil {
		value := serviceModelHealthCheckHTTP{}
		diags.Append(value.FromAPI(ctx, hc.HTTP)...)

		m.HTTP, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	return diags
}

func (m *serviceModelHealthCheck) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type serviceModelHealthCheckHTTP struct {
	Domain      types.String `tfsdk:"domain"`
	Path        types.String `tfsdk:"path"`
	Response    types.String `tfsdk:"response"`
	TLS         types.Bool   `tfsdk:"tls"`
	StatusCodes types.List   `tfsdk:"status_codes"`
}

func (m *serviceModelHealthCheckHTTP) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":       types.StringType,
		"path":         types.StringType,
		"response":     types.StringType,
		"tls":          types.BoolType,
		"status_codes": types.ListType{ElemType: types.StringType},
	}
}

func (m *serviceModelHealthCheckHTTP) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.LoadBalancerServiceHealthCheckHTTP] = &serviceModelHealthCheckHTTP{}
var _ util.ModelToTerraform[types.Object] = &serviceModelHealthCheckHTTP{}

func (m *serviceModelHealthCheckHTTP) FromAPI(_ context.Context, hc *hcloud.LoadBalancerServiceHealthCheckHTTP) diag.Diagnostics {
	var diags, newDiags diag.Diagnostics

	var statusCodes []attr.Value
	for _, code := range hc.StatusCodes {
		statusCodes = append(statusCodes, types.StringValue(code))
	}

	m.Domain = types.StringValue(hc.Domain)
	m.Path = types.StringValue(hc.Path)
	m.Response = types.StringValue(hc.Response)
	m.TLS = types.BoolValue(hc.TLS)
	m.StatusCodes, newDiags = types.ListValue(types.StringType, statusCodes)
	diags = append(diags, newDiags...)

	return diags
}

func (m *serviceModelHealthCheckHTTP) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
