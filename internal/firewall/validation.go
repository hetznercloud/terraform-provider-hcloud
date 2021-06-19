package firewall

import (
	"net"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

func validateIPDiag(i interface{}, path cty.Path) diag.Diagnostics {
	ipS := i.(string)
	ip, n, err := net.ParseCIDR(ipS)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if ip.String() != n.IP.String() {
		return diag.Errorf("%s is not the start of the cidr block %s", ipS, n)
	}
	return nil
}
