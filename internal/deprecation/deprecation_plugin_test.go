package deprecation

import (
	"context"
	"testing"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/stretchr/testify/assert"
)

func TestNewDeprecationModel(t *testing.T) {
	ctx := context.Background()
	fakeTime := time.Date(2024, time.September, 6, 12, 0, 0, 0, time.UTC)

	{
		data, diags := NewDeprecationModel(ctx, hcloud.ServerType{})
		assert.Equal(t, diags.HasError(), false)
		assert.Equal(t, data.IsDeprecated.ValueBool(), false)
		assert.Equal(t, data.DeprecationAnnounced.ValueString(), "")
		assert.Equal(t, data.UnavailableAfter.ValueString(), "")
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
		assert.Equal(t, diags.HasError(), false)
		assert.Equal(t, data.IsDeprecated.ValueBool(), true)
		assert.Equal(t, data.DeprecationAnnounced.ValueString(), "2024-09-06T12:00:00Z")
		assert.Equal(t, data.UnavailableAfter.ValueString(), "2024-09-06T12:00:00Z")
	}
}
