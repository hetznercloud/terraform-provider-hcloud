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
}

// WaitForAction uses ProgressWatcher to wait for the completion of a.
func WaitForAction(ctx context.Context, w ProgressWatcher, a *hcloud.Action) error {
	const op = "hcclient/WaitForAction"

	progC, errC := w.WatchProgress(ctx, a)
	for {
		select {
		case <-progC:
			// only consume the progress channel to avoid blocks in hcloud-go.
			// Once WatchProgress closes both channels once it is done.
		case err := <-errC:
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			// hcloud-go closes the channel once it is finished watching.
			return nil
		}
	}
}
