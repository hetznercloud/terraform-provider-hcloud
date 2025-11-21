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

	detail.WriteString("An API action for the resource failed.\n\n")
	if action.ErrorMessage != "" {
		detail.WriteString(action.ErrorMessage + "\n\n")
	}
	detail.WriteString(actionDetail(action))

	return diag.NewErrorDiagnostic("Action failed", detail.String())
}

func ActionWaitTimeoutDiagnostic(actions ...*hcloud.Action) diag.Diagnostic {
	descriptions := make([]string, 0, len(actions))
	for _, action := range actions {
		descriptions = append(descriptions, fmt.Sprintf("- Command: %s | ID: %d | Progress: %d%% | Resources: %s", action.Command, action.ID, action.Progress, actionResourceDescription(action)))
	}

	return diag.NewErrorDiagnostic(
		"Timeout while waiting on action(s)",
		fmt.Sprintf("The request was cancelled while we were waiting on actions to complete.\n\n"+
			"Actions that are still running:\n%s\n", strings.Join(descriptions, "\n")))
}

func actionDetail(action *hcloud.Action) string {
	return fmt.Sprintf(`Error code: %s
Command: %s
ID: %d
Resources: %s`,
		action.ErrorCode,
		action.Command,
		action.ID,
		actionResourceDescription(action),
	)
}

func actionResourceDescription(action *hcloud.Action) string {
	resources := make([]string, 0, len(action.Resources))
	for _, resource := range action.Resources {
		resources = append(resources, fmt.Sprintf("%s: %d", resource.Type, resource.ID))
	}

	return strings.Join(resources, ", ")
}
