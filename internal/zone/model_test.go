package zone

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	t.Run("primary", func(t *testing.T) {
		ctx := context.Background()
		in := &hcloud.Zone{
			ID:                 1234,
			Name:               "example.com",
			Mode:               hcloud.ZoneModePrimary,
			Labels:             map[string]string{"key": "value"},
			Protection:         hcloud.ZoneProtection{Delete: false},
			TTL:                10800,
			PrimaryNameservers: []hcloud.ZonePrimaryNameserver{},
			Registrar:          hcloud.ZoneRegistrarUnknown,
			Status:             hcloud.ZoneStatusOk,
			RecordCount:        2,
			AuthoritativeNameservers: hcloud.ZoneAuthoritativeNameservers{
				Assigned: []string{"hydrogen.ns.hetzner.com."},
			},
		}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, int64(1234), o.ID.ValueInt64())
		assert.Equal(t, "example.com", o.Name.ValueString())
		assert.Equal(t, "primary", o.Mode.ValueString())
		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"key": "value"}, labels)
	})

	t.Run("secondary", func(t *testing.T) {
		ctx := context.Background()
		in := &hcloud.Zone{
			ID:         1234,
			Name:       "example.com",
			Mode:       hcloud.ZoneModeSecondary,
			Labels:     map[string]string{"key": "value"},
			Protection: hcloud.ZoneProtection{Delete: false},
			TTL:        10800,
			PrimaryNameservers: []hcloud.ZonePrimaryNameserver{
				{Address: "201.34.56.23"},
				{Address: "201.34.56.24", Port: 53},
				{Address: "201.34.56.25", Port: 53, TSIGAlgorithm: "hmac-sha256", TSIGKey: "2bbd1c5d83655d6866aa06b9bc03bf815b631b5349258b19226b8d74339ebea8"},
			},
			Registrar:   hcloud.ZoneRegistrarUnknown,
			Status:      hcloud.ZoneStatusOk,
			RecordCount: 2,
			AuthoritativeNameservers: hcloud.ZoneAuthoritativeNameservers{
				Assigned: []string{"hydrogen.ns.hetzner.com."},
			},
		}
		o := &model{}
		diags := o.FromAPI(ctx, in)
		assert.False(t, diags.HasError())
		assert.Equal(t, int64(1234), o.ID.ValueInt64())
	})
}
