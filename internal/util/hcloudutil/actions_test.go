package hcloudutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type mockActionWaiter struct {
	t       *testing.T
	err     error
	actions []*hcloud.Action
}

func (a *mockActionWaiter) WaitForFunc(_ context.Context, handleUpdate func(update *hcloud.Action) error, actions ...*hcloud.Action) error {
	require.Equal(a.t, a.actions, actions)

	if a.err != nil {
		return a.err
	}

	for _, action := range a.actions {
		require.NoError(a.t, handleUpdate(action))
	}

	return nil
}

func TestSettleActions(t *testing.T) {
	failedAction1 := &hcloud.Action{
		ID:           1337,
		Status:       hcloud.ActionStatusError,
		Command:      "do_thing",
		ErrorCode:    "failed_thing",
		ErrorMessage: "Thing failed",
	}

	failedAction2 := &hcloud.Action{
		ID:           1338,
		Status:       hcloud.ActionStatusError,
		Command:      "create_server",
		ErrorCode:    "spooky_error",
		ErrorMessage: "Something spooky happened",
	}

	successAction1 := &hcloud.Action{
		ID:      1,
		Status:  hcloud.ActionStatusSuccess,
		Command: "delete_server",
	}

	successAction2 := &hcloud.Action{
		ID:      2,
		Status:  hcloud.ActionStatusSuccess,
		Command: "delete_server",
	}

	for _, tc := range []struct {
		name     string
		mock     *mockActionWaiter
		expected diag.Diagnostics
	}{
		{
			name: "success when all actions are successful",
			mock: &mockActionWaiter{
				actions: []*hcloud.Action{successAction1, successAction2},
			},
			expected: nil,
		},
		{
			name: "API Error Diagnostics when ActionClient.WaitForFunc fails",
			mock: &mockActionWaiter{
				actions: []*hcloud.Action{successAction1, successAction2},
				err:     hcloud.Error{Code: "server_error", Details: "Internal server error"},
			},
			expected: APIErrorDiagnostics(hcloud.Error{Code: "server_error", Details: "Internal server error"}),
		},
		{
			name: "complete diagnostics with failed actions",
			mock: &mockActionWaiter{
				actions: []*hcloud.Action{failedAction1, successAction1, failedAction2, successAction2},
			},
			expected: diag.Diagnostics{ActionErrorDiagnostic(failedAction1), ActionErrorDiagnostic(failedAction2)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock.t = t
			diags := SettleActions(t.Context(), tc.mock, tc.mock.actions...)

			assert.Equal(t, tc.expected, diags)
		})
	}
}

func TestActionErrorDiagnostic(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		action   *hcloud.Action
		expected diag.Diagnostic
	}{
		{
			name: "action with single resource",
			action: &hcloud.Action{
				ID:           int64(1337),
				Status:       hcloud.ActionStatusError,
				Command:      "create_server",
				ErrorCode:    "spooky_error",
				ErrorMessage: "Something spooky happened",
				Resources: []*hcloud.ActionResource{
					{
						ID:   int64(42),
						Type: hcloud.ActionResourceTypeServer,
					},
				},
			},
			expected: diag.NewErrorDiagnostic("Action failed", `An API action for the resource failed.

Something spooky happened

Error code: spooky_error
Command: create_server
ID: 1337
Resources: server: 42
`),
		},
		{
			name: "action with multiple resources",
			action: &hcloud.Action{
				ID:           int64(1338),
				Status:       hcloud.ActionStatusError,
				Command:      "attach_floating_ip",
				ErrorCode:    "server_error",
				ErrorMessage: "Unexpected server error",
				Resources: []*hcloud.ActionResource{
					{
						ID:   int64(42),
						Type: hcloud.ActionResourceTypeServer,
					}, {
						ID:   int64(7),
						Type: hcloud.ActionResourceTypeFloatingIP,
					},
				},
			},
			expected: diag.NewErrorDiagnostic("Action failed", `An API action for the resource failed.

Unexpected server error

Error code: server_error
Command: attach_floating_ip
ID: 1338
Resources: server: 42, floating_ip: 7
`),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, ActionErrorDiagnostic(testCase.action))
		})
	}
}
