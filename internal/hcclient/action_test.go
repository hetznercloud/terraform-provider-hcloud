package hcclient_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

func TestWaitForAction(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "successfully wait for action",
		},
		{
			name: "action fails with error",
			err:  errors.New("some error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var a hcloud.Action

			w := hcclient.NewMockProgressWatcher(t)
			ctx := context.Background()
			done := w.MockWatchOverallProgress(ctx, []*hcloud.Action{&a}, tt.err)
			err := hcclient.WaitForActions(ctx, w, []*hcloud.Action{&a})
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error %v; got %v", tt.err, err)
			}

			select {
			case <-done:
				break
			case <-time.After(1000 * time.Millisecond):
				t.Errorf("MockProgressWatcher failed to terminate")
			}

			w.AssertExpectations(t)
		})
	}
}

func TestWaitForActions(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "successfully wait for action",
		},
		{
			name: "action fails with error",
			err:  errors.New("some error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var a hcloud.Action

			w := hcclient.NewMockProgressWatcher(t)
			ctx := context.Background()
			done := w.MockWatchOverallProgress(ctx, []*hcloud.Action{&a}, tt.err)
			err := hcclient.WaitForActions(ctx, w, []*hcloud.Action{&a})
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error %v; got %v", tt.err, err)
			}

			select {
			case <-done:
				break
			case <-time.After(1000 * time.Millisecond):
				t.Errorf("MockWatchOverallProgress failed to terminate")
			}

			w.AssertExpectations(t)
		})
	}
}
