package zonerecord

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
				{Value: "201.45.3.46", Comment: "some web server"},
			},
		}
		o := &model{}
		diags := o.FromAPI(ctx, in)
		assert.False(t, diags.HasError())
		assert.Equal(t, "www", o.Name.ValueString())
		assert.Equal(t, "A", o.Type.ValueString())
		assert.Equal(t, "201.45.3.46", o.Value.ValueString())
		assert.Equal(t, "some web server", o.Comment.ValueString())
	})
}
