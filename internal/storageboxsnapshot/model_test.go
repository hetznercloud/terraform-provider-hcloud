package storageboxsnapshot

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := t.Context()

	in := &hcloud.StorageBoxSnapshot{
		StorageBox:  &hcloud.StorageBox{ID: 1337},
		ID:          1234,
		Name:        "backups",
		Description: "my perfect little backups",
		IsAutomatic: true,
		Labels:      map[string]string{"key": "value"},
		Stats: hcloud.StorageBoxSnapshotStats{
			Size:           1,
			SizeFilesystem: 2,
		},
	}

	o := &dataSourceCommonModel{} // Embeds model
	assert.Nil(t, o.FromAPI(ctx, in))
	assert.Equal(t, int64(1337), o.StorageBox.ValueInt64())
	assert.Equal(t, int64(1234), o.ID.ValueInt64())
	assert.Equal(t, "backups", o.Name.ValueString())
	assert.Equal(t, "my perfect little backups", o.Description.ValueString())
	assert.Equal(t, true, o.IsAutomatic.ValueBool())

	{
		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"key": "value"}, labels)
	}

	{
		m := &modelStats{}
		assert.Nil(t, m.FromTerraform(ctx, o.Stats))

		assert.Equal(t, int64(1), m.Size.ValueInt64())
		assert.Equal(t, int64(2), m.SizeFilesystem.ValueInt64())
	}
}
