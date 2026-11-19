package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/mockutil"
)

func TestPublicNetUpdateDecision(t *testing.T) {
	newClient := func(t *testing.T, requests []mockutil.Request) *hcloud.Client {
		server := mockutil.NewServer(t, requests)
		return hcloud.NewClient(
			hcloud.WithEndpoint(server.URL),
			hcloud.WithRetryOpts(hcloud.RetryOpts{BackoffFunc: hcloud.ConstantBackoff(0), MaxRetries: 3}),
			hcloud.WithPollOpts(hcloud.PollOpts{BackoffFunc: hcloud.ConstantBackoff(0)}),
		)
	}

	t.Run("primary ip already assigned", func(t *testing.T) {
		// After an import, public_net is not in the state, so the configured
		// primary ip shows up as added while the server already uses it.
		client := newClient(t, nil)

		diags := publicNetUpdateDecision(context.Background(), client,
			true, true, 42, 0, &hcloud.Server{ID: 1}, 42, hcloud.PrimaryIPTypeIPv4)
		assert.False(t, diags.HasError(), diags)
	})

	t.Run("replace auto generated primary ip", func(t *testing.T) {
		client := newClient(t, []mockutil.Request{
			{
				Method: "POST", Path: "/primary_ips/7/actions/unassign",
				Status:  http.StatusCreated,
				JSONRaw: `{"action": {"id": 10, "status": "running"}}`,
			},
			{
				Method: "GET", Path: "/actions?id=10&page=1&sort=status&sort=id",
				Status:  http.StatusOK,
				JSONRaw: `{"actions": [{"id": 10, "status": "success"}]}`,
			},
			{
				Method: "DELETE", Path: "/primary_ips/7",
				Status: http.StatusNoContent,
			},
			{
				Method: "POST", Path: "/primary_ips/42/actions/assign",
				Status:  http.StatusCreated,
				JSONRaw: `{"action": {"id": 11, "status": "running"}}`,
			},
			{
				Method: "GET", Path: "/actions?id=11&page=1&sort=status&sort=id",
				Status:  http.StatusOK,
				JSONRaw: `{"actions": [{"id": 11, "status": "success"}]}`,
			},
		})

		diags := publicNetUpdateDecision(context.Background(), client,
			true, true, 42, 0, &hcloud.Server{ID: 1}, 7, hcloud.PrimaryIPTypeIPv4)
		assert.False(t, diags.HasError(), diags)
	})
}
