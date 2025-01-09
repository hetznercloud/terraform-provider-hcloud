package firewall

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAttachment_FromResourceData(t *testing.T) {
	tests := []struct {
		name      string
		rawData   map[string]interface{}
		att       attachment
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name: "server_ids and label_selectors present",
			rawData: map[string]interface{}{
				"firewall_id":     4711,
				"server_ids":      []interface{}{1, 2, 3},
				"label_selectors": []interface{}{"key1=value1", "key2=value2"},
			},
			att: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2, 3},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
		},
		{
			name: "only server_ids present",
			rawData: map[string]interface{}{
				"firewall_id": 4712,
				"server_ids":  []interface{}{4, 5, 6},
			},
			att: attachment{
				FirewallID: 4712,
				ServerIDs:  []int{4, 5, 6},
			},
		},
		{
			name: "only label_selectors present",
			rawData: map[string]interface{}{
				"firewall_id":     4713,
				"label_selectors": []interface{}{"key3=value3", "key4=value4"},
			},
			att: attachment{
				FirewallID:     4713,
				LabelSelectors: []string{"key3=value3", "key4=value4"},
			},
		},
		{
			name: "only firewall id present",
			rawData: map[string]interface{}{
				"firewall_id": 4714,
			},
			att: attachment{
				FirewallID: 4714,
			},
			assertErr: func(t assert.TestingT, err error, args ...interface{}) bool {
				return assert.EqualError(t, err, "no resources referenced", args...)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var actual attachment

			data := schema.TestResourceDataRaw(t, AttachmentResource().Schema, tt.rawData)
			err := actual.FromResourceData(data)
			if tt.assertErr != nil {
				tt.assertErr(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.att, actual)
		})
	}
}

func TestAttachment_ToResourceData(t *testing.T) {
	tests := []struct {
		name    string
		rawData map[string]interface{}
		att     attachment
	}{
		{
			name: "server_ids and label_selectors present",
			att: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2, 3},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
		},
		{
			name: "only server_ids present",
			att: attachment{
				FirewallID: 4712,
				ServerIDs:  []int{4, 5, 6},
			},
		},
		{
			name: "only label_selectors present",
			att: attachment{
				FirewallID:     4713,
				LabelSelectors: []string{"key3=value3", "key4=value4"},
			},
		},
		{
			name: "remove pre-existing server_ids",
			rawData: map[string]interface{}{
				"firewall_id": 4714,
				"server_ids":  []interface{}{1, 2, 3},
			},
			att: attachment{
				FirewallID:     4714,
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
		},
		{
			name: "remove pre-existing label_selectors",
			rawData: map[string]interface{}{
				"firewall_id":     4714,
				"label_selectors": []interface{}{"key1=value1", "key2=value2"},
			},
			att: attachment{
				FirewallID: 4714,
				ServerIDs:  []int{1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, AttachmentResource().Schema, tt.rawData)

			tt.att.ToResourceData(data)

			assert.Equal(t, data.Id(), strconv.Itoa(tt.att.FirewallID))
			assert.Equal(t, data.Get("firewall_id"), tt.att.FirewallID)

			srvIDdata, ok := data.GetOk("server_ids")
			if len(tt.att.ServerIDs) > 0 {
				assert.True(t, ok, "expected data to contain server_ids")

				// Need to iterate as the types of the slices don't match: []int vs []interface{}
				for _, id := range tt.att.ServerIDs {
					assert.Contains(t, srvIDdata.(*schema.Set).List(), id)
				}
			} else {
				assert.False(t, ok, "expected no server_ids in data")
			}

			labelSelData, ok := data.GetOk("label_selectors")
			if len(tt.att.LabelSelectors) > 0 {
				assert.True(t, ok, "expected data to contain label_selectors")

				for _, ls := range tt.att.LabelSelectors {
					assert.Contains(t, labelSelData.(*schema.Set).List(), ls)
				}
			} else {
				assert.False(t, ok, "expected no label_selectors in data")
			}
		})
	}
}

func TestAttachment_FromFirewall(t *testing.T) {
	tests := []struct {
		name        string
		fw          *hcloud.Firewall // partial data is enough for this test.
		att         attachment
		assertError assert.ErrorAssertionFunc
	}{
		{
			name: "nothing attached to firewall",
			fw:   &hcloud.Firewall{ID: 4711},
			att:  attachment{FirewallID: 4711},
		},
		{
			name: "only servers attached to firewall",
			fw: &hcloud.Firewall{
				ID: 4712,
				AppliedTo: []hcloud.FirewallResource{
					serverResource(1),
					serverResource(2),
				},
			},
			att: attachment{
				FirewallID: 4712,
				ServerIDs:  []int{1, 2},
			},
		},
		{
			name: "only label selectors attached to firewall",
			fw: &hcloud.Firewall{
				ID: 4713,
				AppliedTo: []hcloud.FirewallResource{
					labelSelectorResource("key1=value1"),
					labelSelectorResource("key2=value2"),
				},
			},
			att: attachment{
				FirewallID:     4713,
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
		},
		{
			name: "invalid attachment type",
			fw: &hcloud.Firewall{
				ID: 4714,
				AppliedTo: []hcloud.FirewallResource{
					{
						Type: hcloud.FirewallResourceType("invalid"),
					},
				},
			},
			assertError: func(t assert.TestingT, err error, args ...interface{}) bool {
				return assert.EqualError(t, err, "invalid firewall resource type: invalid", args...)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			att := attachment{FirewallID: tt.fw.ID}
			err := att.FromFirewall(tt.fw)
			if tt.assertError != nil {
				tt.assertError(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.att, att)
		})
	}
}

func TestAttachment_AllResources(t *testing.T) {
	tests := []struct {
		name string
		att  attachment
		res  []hcloud.FirewallResource
	}{
		{
			name: "no resources attached",
			att:  attachment{FirewallID: 4711},
		},
		{
			name: "servers and label selectors attached",
			att: attachment{
				FirewallID:     4712,
				ServerIDs:      []int{1, 2},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
			res: []hcloud.FirewallResource{
				serverResource(1),
				serverResource(2),
				labelSelectorResource("key1=value1"),
				labelSelectorResource("key2=value2"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.att.AllResources()
			assert.ElementsMatch(t, tt.res, actual)
		})
	}
}

func TestAttachment_DiffResources(t *testing.T) {
	tests := []struct {
		name  string
		att   attachment
		other attachment
		more  []hcloud.FirewallResource
		less  []hcloud.FirewallResource
	}{
		{
			name: "nothing changed",
			att: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2, 3},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
			other: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2, 3},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
		},
		{
			name: "resources in att but not in other",
			att: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2, 3},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
			other: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2},
				LabelSelectors: []string{"key1=value1"},
			},
			more: []hcloud.FirewallResource{
				serverResource(3),
				labelSelectorResource("key2=value2"),
			},
		},
		{
			name: "resources in other but not in att",
			att: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2},
				LabelSelectors: []string{"key1=value1"},
			},
			other: attachment{
				FirewallID:     4711,
				ServerIDs:      []int{1, 2, 3},
				LabelSelectors: []string{"key1=value1", "key2=value2"},
			},
			less: []hcloud.FirewallResource{
				serverResource(3),
				labelSelectorResource("key2=value2"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			less, more := tt.att.DiffResources(tt.other)
			assert.ElementsMatch(t, tt.less, less)
			assert.ElementsMatch(t, tt.more, more)
		})
	}
}
