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
			name: "Invalid IP",
			ip:   "test",
			err:  diag.Diagnostics{diag.Diagnostic{Severity: 0, Summary: "invalid CIDR address: test", Detail: "", AttributePath: cty.Path(nil)}},
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateIPDiag(test.ip, cty.Path{})
			if test.err == nil {
				assert.Nil(t, err)
			}

			if test.err != nil {
				assert.Equal(t, test.err, err)
			}
		})
	}
}

func Test_normalizeIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Valid CIDR (IPv4)",
			input: "192.0.2.0/24",
			want:  "192.0.2.0/24",
		},
		{
			name:  "IP Address (IPv4)",
			input: "192.0.2.31",
			want:  "192.0.2.31/32",
		},
		{
			name:  "Valid CIDR (IPv6)",
			input: "2001:db8:123:4567::/64",
			want:  "2001:db8:123:4567::/64",
		},
		{
			name:  "Unreduced CIDR (IPv6)",
			input: "2001:0db8:0123:4567::0/64",
			want:  "2001:db8:123:4567::/64",
		},
		{
			name:  "Uppercase CIDR (IPv6)",
			input: "2001:DB8:123:4567::/64",
			want:  "2001:db8:123:4567::/64",
		},
		{
			name:  "IP Address (IPv6)",
			input: "2001:db8:123:4567::",
			want:  "2001:db8:123:4567::/128",
		},
		{
			name:  "Unreduced Matching-All CIDR (IPv6)",
			input: "::0/0",
			want:  "::/0",
		},
		{
			name:  "Badly formatted IP returns input",
			input: "foobar",
			want:  "foobar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeIP(tt.input)
			assert.Equalf(t, tt.want, got, "normalizeIP(%v)", tt.input)
		})
	}
}
