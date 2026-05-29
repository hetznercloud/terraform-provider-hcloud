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

func (m *serviceModel) tfAttributesTypesHTTP() map[string]attr.Type {
	return map[string]attr.Type{
		"sticky_sessions": types.BoolType,
		"cookie_name":     types.StringType,
		"cookie_lifetime": types.Int32Type,
		"certificates":    types.ListType{ElemType: types.Int64Type},
		"redirect_http":   types.BoolType,
		"timeout_idle":    types.Int32Type,
	}
}

func (m *serviceModel) tfAttributesTypesHealthCheckHTTP() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":       types.StringType,
		"path":         types.StringType,
		"response":     types.StringType,
		"tls":          types.BoolType,
		"status_codes": types.ListType{ElemType: types.StringType},
	}
}

func (m *serviceModel) tfAttributesTypesHealthCheck() map[string]attr.Type {
	return map[string]attr.Type{
		"protocol": types.StringType,
		"port":     types.Int32Type,
		"interval": types.Int32Type,
		"timeout":  types.Int32Type,
		"retries":  types.Int32Type,
		"http":     types.ObjectType{AttrTypes: m.tfAttributesTypesHealthCheckHTTP()},
	}
}

func (m *serviceModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"load_balancer_id": types.Int64Type,
		"protocol":         types.StringType,
		"listen_port":      types.Int32Type,
		"destination_port": types.Int32Type,
		"proxyprotocol":    types.BoolType,
		"http":             types.ObjectType{AttrTypes: m.tfAttributesTypesHTTP()},
		"health_check":     types.ObjectType{AttrTypes: m.tfAttributesTypesHealthCheck()},
	}
}

func (m *serviceModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.LoadBalancerService] = &serviceModel{}
var _ util.ModelToTerraform[types.Object] = &serviceModel{}

func (m *serviceModel) FromAPI(_ context.Context, hc *hcloud.LoadBalancerService) diag.Diagnostics {
	var httpCertIDs []attr.Value
	for _, cert := range hc.HTTP.Certificates {
		httpCertIDs = append(httpCertIDs, types.Int64Value(cert.ID))
	}

	m.ListenPort = types.Int32Value(util.CastInt32(hc.ListenPort))
	m.DestinationPort = types.Int32Value(util.CastInt32(hc.DestinationPort))
	m.Proxyprotocol = types.BoolValue(hc.Proxyprotocol)
	m.Protocol = types.StringValue(string(hc.Protocol))
	m.HTTP = types.ObjectValueMust(m.tfAttributesTypesHTTP(), map[string]attr.Value{
		"sticky_sessions": types.BoolValue(hc.HTTP.StickySessions),
		"cookie_name":     types.StringValue(hc.HTTP.CookieName),
		"cookie_lifetime": types.Int32Value(util.CastInt32(hc.HTTP.CookieLifetime.Seconds())),
		"certificates":    types.ListValueMust(types.Int64Type, httpCertIDs),
		"redirect_http":   types.BoolValue(hc.HTTP.RedirectHTTP),
		"timeout_idle":    types.Int32Value(util.CastInt32(hc.HTTP.TimeoutIdle.Seconds())),
	})

	healthCheck := map[string]attr.Value{
		"protocol": types.StringValue(string(hc.HealthCheck.Protocol)),
		"port":     types.Int32Value(util.CastInt32(hc.HealthCheck.Port)),
		"interval": types.Int32Value(util.CastInt32(hc.HealthCheck.Interval.Seconds())),
		"timeout":  types.Int32Value(util.CastInt32(hc.HealthCheck.Timeout.Seconds())),
		"retries":  types.Int32Value(util.CastInt32(hc.HealthCheck.Retries)),
	}

	if hc.HealthCheck.HTTP != nil {
		var statusCodes []attr.Value
		for _, code := range hc.HealthCheck.HTTP.StatusCodes {
			statusCodes = append(statusCodes, types.StringValue(code))
		}

		healthCheck["http"] = types.ObjectValueMust(m.tfAttributesTypesHealthCheckHTTP(), map[string]attr.Value{
			"domain":       types.StringValue(hc.HealthCheck.HTTP.Domain),
			"path":         types.StringValue(hc.HealthCheck.HTTP.Path),
			"response":     types.StringValue(hc.HealthCheck.HTTP.Response),
			"tls":          types.BoolValue(hc.HealthCheck.HTTP.TLS),
			"status_codes": types.ListValueMust(types.StringType, statusCodes),
		})
	}

	m.HealthCheck = types.ObjectValueMust(m.tfAttributesTypesHealthCheck(), healthCheck)

	return nil
}

func (m *serviceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
