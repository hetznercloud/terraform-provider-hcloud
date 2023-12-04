package firewall

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestValidateIPDiag(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		err  diag.Diagnostics
	}{
		{
			name: "Valid CIDR (IPv4)",
			ip:   "10.0.0.0/8",
			err:  nil,
		},
		{
			name: "Valid CIDR (IPv6)",
			ip:   "fe80::/128",
			err:  nil,
		},
		{
			name: "Host bit set (IPv4)",
			ip:   "10.0.0.5/8",
			err:  diag.Diagnostics{diag.Diagnostic{Severity: 0, Summary: "10.0.0.5/8 is not the start of the cidr block 10.0.0.0/8", Detail: "", AttributePath: cty.Path(nil)}},
		},
		{
			name: "Host bit set (IPv6)",
			ip:   "fe80::1337/64",
			err:  diag.Diagnostics{diag.Diagnostic{Severity: 0, Summary: "fe80::1337/64 is not the start of the cidr block fe80::/64", Detail: "", AttributePath: cty.Path(nil)}},
		},
		{
			name: "Valid IP (IPv4)",
			ip:   "10.0.0.1",
			err:  nil,
		},
		{
			name: "Valid IP (IPv6)",
			ip:   "fe80::",
			err:  nil,
		},
		{
			name: "Invalid IP (IPv4)",
			ip:   "10.0.0.256",
			err:  diag.Diagnostics{diag.Diagnostic{Severity: 0, Summary: "invalid IP address: 10.0.0.256", Detail: "", AttributePath: cty.Path(nil)}},
		},
		{
			name: "Invalid IP (IPv6)",
			ip:   "fe80::1337::",
			err:  diag.Diagnostics{diag.Diagnostic{Severity: 0, Summary: "invalid IP address: fe80::1337::", Detail: "", AttributePath: cty.Path(nil)}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateIPDiag(test.ip, cty.Path{})
			if test.err == nil {
				assert.Nil(t, err)
			}

			if test.err != nil {
				assert.Equal(t, err, test.err)
			}
		})
	}
}
