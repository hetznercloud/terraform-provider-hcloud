package hcclient

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// ProgressWatcher encapsulates the Hetzner Cloud Action Client's WatchProgress
// method for easier testing.
type ProgressWatcher interface {
	WatchProgress(context.Context, *hcloud.Action) (<-chan int, <-chan error)
	WatchOverallProgress(ctx context.Context, actions []*hcloud.Action) (<-chan int, <-chan error)
}

// WaitForAction uses ProgressWatcher to wait for the completion of a.
func WaitForAction(ctx context.Context, w ProgressWatcher, a *hcloud.Action) error {
	return WaitForActions(ctx, w, []*hcloud.Action{a})
}

// WaitForActions uses ProgressWatcher to wait for the completion of all actions.
func WaitForActions(ctx context.Context, w ProgressWatcher, a []*hcloud.Action) error {
	const op = "hcclient/WaitForActions"

	progC, errC := w.WatchOverallProgress(ctx, a)
	for {
		select {
		case <-progC:
			// only consume the progress channel to avoid blocks in hcloud-go.
			// Once WaitForActions closes both channels once it is done.
		case err := <-errC:
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			// hcloud-go closes the channel once it is finished watching.
			return nil
		}
	}
}
