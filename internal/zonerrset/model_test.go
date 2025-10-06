package zonerrset

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		in := &hcloud.ZoneRRSet{
			ID:         "www/A",
			Name:       "www",
			Type:       hcloud.ZoneRRSetTypeA,
			Labels:     map[string]string{"key": "value"},
			Protection: hcloud.ZoneRRSetProtection{Change: false},
			TTL:        hcloud.Ptr(10800),
			Records: []hcloud.ZoneRRSetRecord{
				{Value: "201.45.3.45"},
				{Value: "201.45.3.46", Comment: "some web server"},
			},
		}
		o := &model{}
		diags := o.FromAPI(ctx, in)
		assert.False(t, diags.HasError())
		assert.Equal(t, "www/A", o.ID.ValueString())
		assert.Equal(t, "www", o.Name.ValueString())
		assert.Equal(t, "A", o.Type.ValueString())
		labels := make(map[string]string)
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"key": "value"}, labels)
	})
}
