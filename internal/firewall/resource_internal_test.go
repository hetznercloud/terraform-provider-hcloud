package firewall

import (
	"net"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestGetRulesFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		rawData  map[string]any
		expected []hcloud.FirewallRule
	}{
		{
			name: "no rules present",
			rawData: map[string]any{
				"name": "test-firewall",
			},
			expected: []hcloud.FirewallRule{},
		},
		{
			name: "empty rules set",
			rawData: map[string]any{
				"name": "test-firewall",
				"rule": []any{},
			},
			expected: []hcloud.FirewallRule{},
		},
		{
			name: "rules present",
			rawData: map[string]any{
				"name": "test-firewall",
				"rule": []any{
					map[string]any{
						"direction":   "in",
						"protocol":    "tcp",
						"port":        "80",
						"description": "allow http in",
						"source_ips":  []any{"0.0.0.0/0", "::/0"},
					},
				},
			},
			expected: []hcloud.FirewallRule{
				{
					Direction:   hcloud.FirewallRuleDirectionIn,
					Protocol:    hcloud.FirewallRuleProtocolTCP,
					Port:        new("80"),
					Description: new("allow http in"),
					SourceIPs: []net.IPNet{
						{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
						{IP: net.ParseIP("::"), Mask: net.CIDRMask(0, 128)},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, Resource().Schema, tt.rawData)
			rules := getRulesFromResourceData(data)
			require.NotNil(t, rules)
			assert.Len(t, rules, len(tt.expected))
			if len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Direction, rules[0].Direction)
				assert.Equal(t, tt.expected[0].Protocol, rules[0].Protocol)
				assert.Equal(t, *tt.expected[0].Port, *rules[0].Port)
				assert.Equal(t, *tt.expected[0].Description, *rules[0].Description)
			}
		})
	}
}
