package hcloudutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type ActionWaiter interface {
	WaitForFunc(ctx context.Context, handleUpdate func(update *hcloud.Action) error, actions ...*hcloud.Action) error
}

func SettleActions(ctx context.Context, client ActionWaiter, actions ...*hcloud.Action) diag.Diagnostics {
	var diags diag.Diagnostics

	failedActions := make([]*hcloud.Action, 0)

	err := client.WaitForFunc(ctx, func(update *hcloud.Action) error {
		switch update.Status {
		case hcloud.ActionStatusError:
			failedActions = append(failedActions, update)
		default:
			// Do nothing
		}

		return nil
	}, actions...)
	if err != nil {
		return APIErrorDiagnostics(err)
	}

	if len(failedActions) > 0 {
		for _, action := range failedActions {
			diags.Append(ActionErrorDiagnostic(action))
		}
	}

	return diags
}

func ActionErrorDiagnostic(action *hcloud.Action) diag.Diagnostic {
	resources := make([]string, 0, len(action.Resources))
	for _, resource := range action.Resources {
		resources = append(resources, fmt.Sprintf("%s: %d", resource.Type, resource.ID))
	}

	return diag.NewErrorDiagnostic("Action failed", fmt.Sprintf(
		"An API action for the resource failed.\n\n"+
			"%s\n\n"+
			"Error code: %s\n"+
			"Command: %s\n"+
			"ID: %d\n"+
			"Resources: %s\n",
		action.ErrorMessage, action.ErrorCode, action.Command, action.ID, strings.Join(resources, ", "),
	))
}
