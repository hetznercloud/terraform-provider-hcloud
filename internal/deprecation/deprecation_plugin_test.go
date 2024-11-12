package deprecation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestNewDeprecationModel(t *testing.T) {
	ctx := context.Background()
	fakeTime := time.Date(2024, time.September, 6, 12, 0, 0, 0, time.UTC)

	{
		data, diags := NewDeprecationModel(ctx, hcloud.ServerType{})
		assert.Equal(t, false, diags.HasError())
		assert.Equal(t, false, data.IsDeprecated.ValueBool())
		assert.Equal(t, false, data.DeprecationAnnounced.IsNull())
		assert.Equal(t, "", data.DeprecationAnnounced.ValueString())
		assert.Equal(t, false, data.UnavailableAfter.IsNull())
		assert.Equal(t, "", data.UnavailableAfter.ValueString())
	}

	{
		data, diags := NewDeprecationModel(ctx, hcloud.ServerType{
			DeprecatableResource: hcloud.DeprecatableResource{
				Deprecation: &hcloud.DeprecationInfo{
					Announced:        fakeTime,
					UnavailableAfter: fakeTime,
				},
			},
		})
		assert.Equal(t, false, diags.HasError())
		assert.Equal(t, true, data.IsDeprecated.ValueBool())
		assert.Equal(t, "2024-09-06T12:00:00Z", data.DeprecationAnnounced.ValueString())
		assert.Equal(t, "2024-09-06T12:00:00Z", data.UnavailableAfter.ValueString())
	}
}
