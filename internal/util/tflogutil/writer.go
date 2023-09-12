package tflogutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Writer wraps the logging interface terraform provides in a [io.Writer] interface
// that hcloud-go expects for its HTTP logs.
type Writer struct {
	// Context holds the tflog writer that is used to write the logs to.
	ctx context.Context
}

func (w *Writer) Write(p []byte) (n int, err error) {
	tflog.Debug(w.ctx, string(p))
	return len(p), nil
}

func NewWriter(ctx context.Context) *Writer {
	return &Writer{
		ctx: ctx,
	}
}
