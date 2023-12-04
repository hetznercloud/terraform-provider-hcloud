package firewall

import (
	"net"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

func validateIPDiag(i interface{}, _ cty.Path) diag.Diagnostics {
	ipS := i.(string)
	if !strings.Contains(ipS, "/") {
		if net.ParseIP(ipS) == nil {
			return diag.Errorf("invalid IP address: %s", ipS)
		}
		if strings.Contains(ipS, ":") {
			ipS += "/64"
		} else if strings.Contains(ipS, ".") {
			ipS += "/32"
		}
	}
	ip, n, err := net.ParseCIDR(ipS)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if ip.String() != n.IP.String() {
		return diag.Errorf("%s is not the start of the cidr block %s", ipS, n)
	}
	return nil
}
