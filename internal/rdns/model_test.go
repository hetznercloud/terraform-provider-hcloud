package rdns

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ip := net.ParseIP("203.0.113.10")
	dnsPtr := "host.example.org"

	t.Run("server", func(t *testing.T) {
		ctx := context.Background()

		in := &hcloud.Server{ID: 1234, Name: "example"}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in, ip, dnsPtr))

		assert.Equal(t, "s-1234-203.0.113.10", o.ID.ValueString())

		assert.Equal(t, "203.0.113.10", o.IPAddress.ValueString())
		assert.Equal(t, dnsPtr, o.DNSPtr.ValueString())

		// assert.True(t, o.ServerID.IsNull())
		assert.Equal(t, int64(1234), o.ServerID.ValueInt64())
		assert.True(t, o.PrimaryIPID.IsNull())
		assert.True(t, o.FloatingIPID.IsNull())
		assert.True(t, o.LoadBalancerID.IsNull())
	})

	t.Run("primary ip", func(t *testing.T) {
		ctx := context.Background()

		in := &hcloud.PrimaryIP{ID: 1234, Name: "example"}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in, ip, dnsPtr))

		assert.Equal(t, "p-1234-203.0.113.10", o.ID.ValueString())

		assert.Equal(t, "203.0.113.10", o.IPAddress.ValueString())
		assert.Equal(t, dnsPtr, o.DNSPtr.ValueString())

		assert.True(t, o.ServerID.IsNull())
		// assert.True(t, o.PrimaryIPID.IsNull())
		assert.Equal(t, int64(1234), o.PrimaryIPID.ValueInt64())
		assert.True(t, o.FloatingIPID.IsNull())
		assert.True(t, o.LoadBalancerID.IsNull())
	})

	t.Run("floating ip", func(t *testing.T) {
		ctx := context.Background()

		in := &hcloud.FloatingIP{ID: 1234, Name: "example"}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in, ip, dnsPtr))

		assert.Equal(t, "f-1234-203.0.113.10", o.ID.ValueString())

		assert.Equal(t, "203.0.113.10", o.IPAddress.ValueString())
		assert.Equal(t, dnsPtr, o.DNSPtr.ValueString())

		assert.True(t, o.ServerID.IsNull())
		assert.True(t, o.PrimaryIPID.IsNull())
		// assert.True(t, o.FloatingIPID.IsNull())
		assert.Equal(t, int64(1234), o.FloatingIPID.ValueInt64())
		assert.True(t, o.LoadBalancerID.IsNull())
	})

	t.Run("load balancer", func(t *testing.T) {
		ctx := context.Background()

		in := &hcloud.LoadBalancer{ID: 1234, Name: "example"}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in, ip, dnsPtr))

		assert.Equal(t, "l-1234-203.0.113.10", o.ID.ValueString())

		assert.Equal(t, "203.0.113.10", o.IPAddress.ValueString())
		assert.Equal(t, dnsPtr, o.DNSPtr.ValueString())

		assert.True(t, o.ServerID.IsNull())
		assert.True(t, o.PrimaryIPID.IsNull())
		assert.True(t, o.FloatingIPID.IsNull())
		// assert.True(t, o.LoadBalancerID.IsNull())
		assert.Equal(t, int64(1234), o.LoadBalancerID.ValueInt64())
	})

}
