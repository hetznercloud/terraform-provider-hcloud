package storagebox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestModel(t *testing.T) {
	ctx := t.Context()

	in := &hcloud.StorageBox{
		ID:             1234,
		Username:       "u1",
		Name:           "backups",
		StorageBoxType: &hcloud.StorageBoxType{Name: "bx11"},
		Location:       &hcloud.Location{Name: "fsn1"},
		AccessSettings: hcloud.StorageBoxAccessSettings{
			ReachableExternally: true,
			SambaEnabled:        false,
			SSHEnabled:          true,
			WebDAVEnabled:       false,
			ZFSEnabled:          true,
		},
		Server:     "u1.your-storagebox.de",
		System:     "FSN1-BX42",
		Labels:     map[string]string{"key": "value"},
		Protection: hcloud.StorageBoxProtection{Delete: true},
		SnapshotPlan: &hcloud.StorageBoxSnapshotPlan{
			MaxSnapshots: 10,
			Minute:       16,
			Hour:         18,
			DayOfWeek:    hcloud.Ptr(time.Weekday(3)),
		},
	}
	o := &commonModel{}
	assert.Nil(t, o.FromAPI(ctx, in))
	assert.Equal(t, int64(1234), o.ID.ValueInt64())
	assert.Equal(t, "backups", o.Name.ValueString())
	assert.Equal(t, "u1", o.Username.ValueString())
	assert.Equal(t, "bx11", o.StorageBoxType.ValueString())
	assert.Equal(t, "fsn1", o.Location.ValueString())

	{
		m := &modelAccessSettings{}
		assert.Nil(t, m.FromTerraform(ctx, o.AccessSettings))

		assert.Equal(t, true, m.ReachableExternally.ValueBool())
		assert.Equal(t, false, m.SambaEnabled.ValueBool())
		assert.Equal(t, true, m.SSHEnabled.ValueBool())
		assert.Equal(t, false, m.WebDAVEnabled.ValueBool())
		assert.Equal(t, true, m.ZFSEnabled.ValueBool())
	}

	assert.Equal(t, "u1.your-storagebox.de", o.Server.ValueString())
	assert.Equal(t, "FSN1-BX42", o.System.ValueString())

	labels := map[string]string{}
	assert.Nil(t, o.Labels.ElementsAs(ctx, &labels, false))
	assert.Equal(t, map[string]string{"key": "value"}, labels)

	assert.Equal(t, true, o.DeleteProtection.ValueBool())

	{
		m := &modelSnapshotPlan{}
		assert.Nil(t, m.FromTerraform(ctx, o.SnapshotPlan))

		assert.Equal(t, int32(10), m.MaxSnapshots.ValueInt32())
		assert.Equal(t, int32(16), m.Minute.ValueInt32())
		assert.Equal(t, int32(18), m.Hour.ValueInt32())
		assert.Equal(t, int32(3), m.DayOfWeek.ValueInt32())
		assert.Equal(t, true, m.DayOfMonth.IsNull())
	}

}
