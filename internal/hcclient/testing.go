package hcclient

import (
	"context"
	"testing"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/stretchr/testify/mock"
)

// MockProgressWatcher provides a mock implementation of the ProgressWatcher
// interface.
type MockProgressWatcher struct {
	mock.Mock
}

// NewMockProgressWatcher creates a new mock progress watcher.
func NewMockProgressWatcher(t *testing.T) *MockProgressWatcher {
	w := &MockProgressWatcher{}
	w.Test(t)
	return w
}

// WatchProgress is a mock implementation of the ProgressWatcher.WatchProgress
// method.
func (m *MockProgressWatcher) WatchProgress(ctx context.Context, a *hcloud.Action) (<-chan int, <-chan error) {
	args := m.Called(ctx, a)
	return testsupport.GetIntChan(args, 0), testsupport.GetErrChan(args, 1)
}

// MockWatchProgress mocks WatchProgress to return a progress and an error
// channel. A Go routine is started, which closes the channels and sends err
// into the error channel if not nil.
func (m *MockProgressWatcher) MockWatchProgress(ctx context.Context, a *hcloud.Action, err error) <-chan struct{} {
	progC := make(chan int)
	errC := make(chan error)
	doneC := make(chan struct{})

	go func() {
		defer close(progC)
		defer close(errC)
		defer close(doneC)

		progC <- 50
		if err != nil {
			errC <- err
			return
		}
		progC <- 100
	}()

	m.On("WatchProgress", ctx, a).Return(progC, errC)
	return doneC
}

// WatchProgress is a mock implementation of the ProgressWatcher.WatchProgress
// method.
func (m *MockProgressWatcher) WatchOverallProgress(ctx context.Context, a []*hcloud.Action) (<-chan int, <-chan error) {
	args := m.Called(ctx, a)
	return testsupport.GetIntChan(args, 0), testsupport.GetErrChan(args, 1)
}

// MockWatchProgress mocks WatchProgress to return a progress and an error
// channel. A Go routine is started, which closes the channels and sends err
// into the error channel if not nil.
func (m *MockProgressWatcher) MockWatchOverallProgress(ctx context.Context, a []*hcloud.Action, err error) <-chan struct{} {
	progC := make(chan int)
	errC := make(chan error)
	doneC := make(chan struct{})

	go func() {
		defer close(progC)
		defer close(errC)
		defer close(doneC)

		progC <- 50
		if err != nil {
			errC <- err
			return
		}
		progC <- 100
	}()

	m.On("WatchOverallProgress", ctx, a).Return(progC, errC)
	return doneC
}
