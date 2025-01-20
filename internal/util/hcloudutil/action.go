package hcloudutil

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// WaitForAction uses [hcloud.ActionWaiter.WaitFor] to wait for the completion of an action.
func WaitForAction(ctx context.Context, c hcloud.ActionWaiter, action *hcloud.Action) error {
	const op = "hcloudutil/WaitForAction"

	if err := c.WaitFor(ctx, action); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// WaitForActions uses [hcloud.ActionWaiter.WaitFor] to wait for the completion of all actions.
func WaitForActions(ctx context.Context, c hcloud.ActionWaiter, actions []*hcloud.Action) error {
	const op = "hcloudutil/WaitForActions"

	if err := c.WaitFor(ctx, actions...); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
