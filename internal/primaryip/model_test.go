package primaryip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := t.Context()

	t.Run("ipv4", func(t *testing.T) {
		in := &hcloud.PrimaryIP{
			ID:           42,
			Name:         "primary-ip",
			Type:         hcloud.PrimaryIPTypeIPv4,
			IP:           net.ParseIP("131.232.99.42"),
			Network:      nil,
			Location:     &hcloud.Location{Name: "fsn1"},
			AssigneeID:   0,
			AssigneeType: "server",
			AutoDelete:   false,
			Labels:       map[string]string{"key": "value"},
			Protection:   hcloud.PrimaryIPProtection{Delete: true},
		}

		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, int64(42), o.ID.ValueInt64())
		assert.Equal(t, "primary-ip", o.Name.ValueString())
		assert.Equal(t, "ipv4", o.Type.ValueString())
		assert.Equal(t, "131.232.99.42", o.IPAddress.ValueString())
		assert.True(t, o.IPNetwork.IsNull())
		assert.Equal(t, "fsn1", o.Location.ValueString())
		assert.True(t, o.Datacenter.IsNull())
		assert.Equal(t, int64(0), o.AssigneeID.ValueInt64())
		assert.Equal(t, "server", o.AssigneeType.ValueString())
		assert.Equal(t, false, o.AutoDelete.ValueBool())

		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"key": "value"}, labels)

		assert.Equal(t, true, o.DeleteProtection.ValueBool())
	})

	t.Run("ipv6", func(t *testing.T) {
		ip, network, err := net.ParseCIDR("2001:db8::/64")
		require.NoError(t, err)

		in := &hcloud.PrimaryIP{
			ID:           42,
			Name:         "primary-ip",
			Type:         hcloud.PrimaryIPTypeIPv6,
			IP:           ip,
			Network:      network,
			Location:     &hcloud.Location{Name: "fsn1"},
			AssigneeID:   0,
			AssigneeType: "server",
			AutoDelete:   false,
			Labels:       map[string]string{"key": "value"},
			Protection:   hcloud.PrimaryIPProtection{Delete: true},
		}

		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, int64(42), o.ID.ValueInt64())
		assert.Equal(t, "primary-ip", o.Name.ValueString())
		assert.Equal(t, "ipv6", o.Type.ValueString())
		assert.Equal(t, "2001:db8::", o.IPAddress.ValueString())
		assert.Equal(t, "2001:db8::/64", o.IPNetwork.ValueString())
		assert.Equal(t, "fsn1", o.Location.ValueString())
		assert.True(t, o.Datacenter.IsNull())
		assert.Equal(t, int64(0), o.AssigneeID.ValueInt64())
		assert.Equal(t, "server", o.AssigneeType.ValueString())
		assert.Equal(t, false, o.AutoDelete.ValueBool())

		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"key": "value"}, labels)

		assert.Equal(t, true, o.DeleteProtection.ValueBool())
	})

	t.Run("ipv4 with datacenter", func(t *testing.T) {
		in := &hcloud.PrimaryIP{
			ID:           42,
			Name:         "primary-ip",
			Type:         hcloud.PrimaryIPTypeIPv4,
			IP:           net.ParseIP("131.232.99.42"),
			Network:      nil,
			Datacenter:   &hcloud.Datacenter{Name: "fsn1-dc14"},
			Location:     &hcloud.Location{Name: "fsn1"},
			AssigneeID:   0,
			AssigneeType: "server",
			AutoDelete:   false,
			Labels:       map[string]string{"key": "value"},
			Protection:   hcloud.PrimaryIPProtection{Delete: true},
		}

		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, "fsn1-dc14", o.Datacenter.ValueString())
	})
}
