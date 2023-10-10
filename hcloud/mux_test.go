package hcloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
)

func TestMuxedProviderSchema(t *testing.T) {
	providerFactory, err := GetMuxedProvider(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	resp, err := providerFactory().GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, resp.Diagnostics, 0)
}
