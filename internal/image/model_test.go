package image

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := t.Context()

	t.Run("system deprecated", func(t *testing.T) {
		in := &hcloud.Image{
			ID:           1234,
			Type:         hcloud.ImageTypeSystem,
			Name:         "debian-13",
			Description:  "",
			Labels:       map[string]string{},
			OSFlavor:     "debian",
			OSVersion:    "trixie",
			Architecture: hcloud.ArchitectureX86,
			RapidDeploy:  true,
			Created:      time.Date(2026, 5, 1, 17, 0, 0, 0, time.UTC),
			Deprecated:   time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC),
		}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, int64(1234), o.ID.ValueInt64())
		assert.Equal(t, "system", o.Type.ValueString())
		assert.Equal(t, "debian-13", o.Name.ValueString())
		assert.Equal(t, "", o.Description.ValueString())

		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{}, labels)

		assert.Equal(t, "debian", o.OSFlavor.ValueString())
		assert.Equal(t, "trixie", o.OSVersion.ValueString())
		assert.Equal(t, "x86", o.Architecture.ValueString())
		assert.Equal(t, true, o.RapidDeploy.ValueBool())
		assert.Equal(t, "2026-05-01T17:00:00Z", o.Created.ValueString())
		assert.Equal(t, "2026-05-20T10:00:00Z", o.Deprecated.ValueString())
	})

	t.Run("snapshot", func(t *testing.T) {
		in := &hcloud.Image{
			ID:           1234,
			Type:         hcloud.ImageTypeSnapshot,
			Name:         "",
			Description:  "snapshot-1234",
			Labels:       map[string]string{"env": "prod"},
			OSFlavor:     "debian",
			OSVersion:    "trixie",
			Architecture: hcloud.ArchitectureX86,
			RapidDeploy:  false,
			Created:      time.Date(2026, 5, 1, 17, 0, 0, 0, time.UTC),
		}
		o := &model{}
		assert.Nil(t, o.FromAPI(ctx, in))
		assert.Equal(t, int64(1234), o.ID.ValueInt64())
		assert.Equal(t, "snapshot", o.Type.ValueString())
		assert.Equal(t, "", o.Name.ValueString())
		assert.Equal(t, "snapshot-1234", o.Description.ValueString())

		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"env": "prod"}, labels)

		assert.Equal(t, "debian", o.OSFlavor.ValueString())
		assert.Equal(t, "trixie", o.OSVersion.ValueString())
		assert.Equal(t, "x86", o.Architecture.ValueString())
		assert.Equal(t, false, o.RapidDeploy.ValueBool())
		assert.Equal(t, "2026-05-01T17:00:00Z", o.Created.ValueString())
		assert.True(t, o.Deprecated.IsNull())
	})
}
