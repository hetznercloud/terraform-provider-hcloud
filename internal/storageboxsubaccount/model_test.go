package storageboxsubaccount

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := t.Context()

	in := &hcloud.StorageBoxSubaccount{
		ID:            1234,
		Username:      "u1",
		HomeDirectory: "foo/bar",
		Server:        "u1-sub1.your-storagebox.de",
		AccessSettings: &hcloud.StorageBoxSubaccountAccessSettings{
			ReachableExternally: true,
			SambaEnabled:        false,
			SSHEnabled:          true,
			WebDAVEnabled:       false,
			Readonly:            true,
		},
		Description: "Baz",
		Labels:      map[string]string{"key": "value"},
		StorageBox:  &hcloud.StorageBox{ID: 1337},
	}

	o := &model{}
	assert.Nil(t, o.FromAPI(ctx, in))
	assert.Equal(t, int64(1337), o.StorageBoxID.ValueInt64())
	assert.Equal(t, int64(1234), o.ID.ValueInt64())
	assert.Equal(t, "u1", o.Username.ValueString())
	assert.Equal(t, "foo/bar", o.HomeDirectory.ValueString())
	assert.Equal(t, "u1-sub1.your-storagebox.de", o.Server.ValueString())
	assert.Equal(t, "Baz", o.Description.ValueString())

	{
		labels := map[string]string{}
		assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
		assert.Equal(t, map[string]string{"key": "value"}, labels)
	}

	{
		m := &modelAccessSettings{}
		assert.Nil(t, m.FromTerraform(ctx, o.AccessSettings))

		assert.Equal(t, true, m.ReachableExternally.ValueBool())
		assert.Equal(t, false, m.SambaEnabled.ValueBool())
		assert.Equal(t, true, m.SSHEnabled.ValueBool())
		assert.Equal(t, false, m.WebDAVEnabled.ValueBool())
		assert.Equal(t, true, m.Readonly.ValueBool())
	}
}
