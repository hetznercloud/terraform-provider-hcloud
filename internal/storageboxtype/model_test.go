package storageboxtype

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := context.Background()
	in := &hcloud.StorageBoxType{
		ID:                     1333,
		Name:                   "bx11",
		Description:            "BX11",
		SnapshotLimit:          hcloud.Ptr(10),
		AutomaticSnapshotLimit: hcloud.Ptr(11),
		SubaccountsLimit:       100,
		Size:                   1 * 1024 * 1024 * 1024 * 1024,
		DeprecatableResource: hcloud.DeprecatableResource{
			Deprecation: &hcloud.DeprecationInfo{
				Announced:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UnavailableAfter: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	o := &model{}
	assert.Nil(t, o.FromAPI(ctx, in))
	assert.Equal(t, int64(1333), o.ID.ValueInt64())
	assert.Equal(t, "bx11", o.Name.ValueString())
	assert.Equal(t, "BX11", o.Description.ValueString())
	assert.Equal(t, int64(10), o.SnapshotLimit.ValueInt64())
	assert.Equal(t, int64(11), o.AutomaticSnapshotLimit.ValueInt64())
	assert.Equal(t, int64(100), o.SubaccountsLimit.ValueInt64())
	assert.Equal(t, int64(1*1024*1024*1024*1024), o.Size.ValueInt64())

	assert.True(t, o.IsDeprecated.ValueBool())
	assert.Equal(t, "2025-01-01T00:00:00Z", o.DeprecationAnnounced.ValueString())
	assert.Equal(t, "2025-04-01T00:00:00Z", o.UnavailableAfter.ValueString())

}

func Test_intPtrToInt64Ptr(t *testing.T) {
	tests := []struct {
		arg  *int
		want *int64
	}{
		{arg: hcloud.Ptr(0), want: hcloud.Ptr(int64(0))},
		{arg: hcloud.Ptr(1337), want: hcloud.Ptr(int64(1337))},
		{arg: nil, want: nil},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert.Equal(t, tt.want, intPtrToInt64Ptr(tt.arg))
		})
	}
}
