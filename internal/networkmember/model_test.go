package networkmember

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := t.Context()

	t.Run("minimal", func(t *testing.T) {
		in := &hcloud.NetworkMember{
			ID:       42,
			Type:     hcloud.NetworkMemberTypeServer,
			Subnet:   &net.IPNet{IP: net.IPv4(10, 0, 1, 0), Mask: net.CIDRMask(24, 32)},
			IP:       net.IPv4(10, 0, 1, 2),
			AliasIPs: []net.IP{net.IPv4(10, 0, 1, 5)},
			Status:   hcloud.NetworkMemberStatusOK,
		}

		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, int64(42), o.ID.ValueInt64())
		assert.Equal(t, "server", o.Type.ValueString())
		assert.Equal(t, "10.0.1.0/24", o.Subnet.ValueString())
		assert.Equal(t, "10.0.1.2", o.IP.ValueString())
		{
			elements := []string{}
			assert.Nil(t, o.AliasIPs.ElementsAs(ctx, &elements, false))
			assert.Equal(t, []string{"10.0.1.5"}, elements)
		}
		assert.Equal(t, "ok", o.Status.ValueString())
	})
}
