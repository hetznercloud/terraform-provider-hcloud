package firewall

import (
	"net"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

var (
	// These masks are also used in the Cloud Console when a user enters an IP without range.
	defaultMaskIPv4 = net.CIDRMask(32, 32)
	defaultMaskIPv6 = net.CIDRMask(128, 128)
)

func validateIPDiag(i interface{}, _ cty.Path) diag.Diagnostics {
	i = normalizeIP(i)

	ipS := i.(string)
	ip, n, err := net.ParseCIDR(ipS)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if ip.String() != n.IP.String() {
		return diag.Errorf("%s is not the start of the cidr block %s", ipS, n)
	}
	return nil
}

// normalizeIP implements two closely related functions:
//  1. It normalizes an IP address or CIDR block to a CIDR block. To allow users to specify the IP directly.
//  2. The API modifies CIDRs to lower case and IPv6 to its minimal form. This function does the same to
//     have clean diffs, even if the user input does not match the desired format by the API.
func normalizeIP(i interface{}) string {
	input := i.(string)

	ip, ipnet, err := net.ParseCIDR(input)
	if err == nil {
		// net.ParseCIDR removes any set host bits. We want to show an error to the user instead,
		// to avoid making any assumptions about their intent.
		// By setting the parse IP in the ipnet, the returned string will be the same as the input, only normalized & lower cased.
		ipnet.IP = ip
	} else {
		ip = net.ParseIP(input)
		if ip == nil {
			// No CIDR or IP, just return the input string
			return input
		}
		if ip.To4() != nil {
			// IPv4

			ipnet = &net.IPNet{
				IP:   ip,
				Mask: defaultMaskIPv4,
			}
		} else {
			// If To4 returns nil, IP is not IPv4 => IPv6
			ipnet = &net.IPNet{
				IP:   ip,
				Mask: defaultMaskIPv6,
			}
		}
	}

	return ipnet.String()
}
