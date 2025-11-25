package hcloudutil

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type ActionWaiter interface {
	WaitForFunc(ctx context.Context, handleUpdate func(update *hcloud.Action) error, actions ...*hcloud.Action) error
}

func SettleActions(ctx context.Context, client ActionWaiter, actions ...*hcloud.Action) (diags diag.Diagnostics) {
	runningActions := make([]*hcloud.Action, len(actions))
	copy(runningActions, actions)
	failedActions := make([]*hcloud.Action, 0)

	// Make sure we always report failed actions, even if an API calls fails for other reasons and we return early
	defer func() {
		if len(failedActions) > 0 {
			for _, action := range failedActions {
				diags.Append(ActionErrorDiagnostic(action))
			}
		}
	}()

	err := client.WaitForFunc(ctx, func(update *hcloud.Action) error {
		if update.Status == hcloud.ActionStatusSuccess || update.Status == hcloud.ActionStatusError {
			// Remove from runningActions
			runningActions = slices.DeleteFunc(runningActions, func(action *hcloud.Action) bool {
				return action.ID == update.ID
			})

			if update.Status == hcloud.ActionStatusError {
				failedActions = append(failedActions, update)
			}
		}

		return nil
	}, actions...)
	if err != nil {
		if (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) && len(runningActions) > 0 {
			diags.Append(ActionWaitTimeoutDiagnostic(runningActions...))
			return
		}

		diags.Append(APIErrorDiagnostics(err)...)
		return
	}

	return
}

func ActionErrorDiagnostic(action *hcloud.Action) diag.Diagnostic {
	detail := strings.Builder{}

	detail.WriteString("An API action for the resource failed.")
	if action.ErrorMessage != "" {
		fmt.Fprint(&detail, "\n\n")
		fmt.Fprint(&detail, action.ErrorMessage)
	}
	fmt.Fprint(&detail, "\n\n")
	fmt.Fprintf(&detail, "Error code: %s\n", action.ErrorCode)
	fmt.Fprintf(&detail, "Command: %s\n", action.Command)
	fmt.Fprintf(&detail, "ID: %d\n", action.ID)
	fmt.Fprintf(&detail, "Resources: %s", actionResourceDescription(action))

	return diag.NewErrorDiagnostic("Action failed", detail.String())
}

func ActionWaitTimeoutDiagnostic(actions ...*hcloud.Action) diag.Diagnostic {
	detail := strings.Builder{}

	detail.WriteString("The request was cancelled while we were waiting on action(s) to complete.")
	for _, action := range actions {
		fmt.Fprint(&detail, "\n\n")
		fmt.Fprintf(&detail, "- Command: %s\n", action.Command)
		fmt.Fprintf(&detail, "  ID: %d\n", action.ID)
		fmt.Fprintf(&detail, "  Progress: %d%%\n", action.Progress)
		fmt.Fprintf(&detail, "  Resources: %s", actionResourceDescription(action))
	}

	return diag.NewErrorDiagnostic("Timeout while waiting on action(s)", detail.String())
}

func actionResourceDescription(action *hcloud.Action) string {
	resources := make([]string, 0, len(action.Resources))
	for _, resource := range action.Resources {
		resources = append(resources, fmt.Sprintf("%s: %d", resource.Type, resource.ID))
	}

	return strings.Join(resources, ", ")
}
