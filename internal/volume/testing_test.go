package volume_test

import (
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
)

// VolumeRData is a resource for use in volume related test.
func VolumeRData() *volume.RData {
	return &volume.RData{
		Name:         "basic-volume",
		LocationName: teste2e.TestLocationName,
		Size:         10,
	}
}
