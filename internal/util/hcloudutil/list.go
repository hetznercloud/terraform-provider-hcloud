package hcloudutil

import (
	"fmt"
	"io"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type WithValues interface {
	Values() url.Values
}

type Details struct {
	resourceName string
	params       url.Values
	usingName    string
	usingValue   any
}

type DetailsOption func(o *Details)

func NewDetails(opts ...DetailsOption) *Details {
	o := &Details{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithResourceName adds the resource name to the details.
func WithResourceName(name string) DetailsOption {
	return func(o *Details) {
		o.resourceName = name
	}
}

// WithUsing adds a hint about which data in particular was used for the query.
// For example `label selector` or `name`.
func WithUsing(name string, value any) DetailsOption {
	return func(o *Details) {
		o.usingName = name
		o.usingValue = value
	}
}

// WithListOpts adds the query params to the details.
func WithListOpts(opts WithValues) DetailsOption {
	return func(o *Details) {
		o.params = opts.Values()
	}
}

func (d *Details) writeUsing(w io.Writer) {
	if d.usingName != "" {
		fmt.Fprintf(w, " using %s", d.usingName)
		if d.usingValue == nil {
			fmt.Fprintf(w, ".")
		} else {
			fmt.Fprintf(w, ": %v", d.usingValue)
		}
	}
}

func (d *Details) writeParams(w io.Writer) {
	if len(d.params) == 0 {
		return
	}

	fmt.Fprint(w, "\n\n")
	for _, key := range slices.Sorted(maps.Keys(d.params)) {
		if len(d.params[key]) == 1 {
			fmt.Fprintf(w, "%s=%s\n", key, d.params[key][0])
		} else {
			fmt.Fprintf(w, "%s=%v\n", key, d.params[key])
		}
	}
}

func (d *Details) NotFound() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Resource (%s) was not found", d.resourceName)
	d.writeUsing(b)
	d.writeParams(b)
	return b.String()
}

func (d *Details) MoreThanOne() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Found more than one resource (%s)", d.resourceName)
	d.writeUsing(b)
	d.writeParams(b)
	return b.String()
}

// GetFirst verifies that at least one item is present in the list and returns the first item.
func GetFirst[T any](items []*T, details ...DetailsOption) (*T, diag.Diagnostic) {
	if len(items) == 0 {
		return nil, diag.NewErrorDiagnostic("Resource not found", NewDetails(details...).NotFound())
	}
	return items[0], nil
}

// GetOne verifies that exactly one item is present in the list and returns that item.
func GetOne[T any](items []*T, details ...DetailsOption) (*T, diag.Diagnostic) {
	if len(items) == 0 {
		return nil, diag.NewErrorDiagnostic("Resource not found", NewDetails(details...).NotFound())
	}
	if len(items) > 1 {
		return nil, diag.NewErrorDiagnostic("Found more than one resource", NewDetails(details...).MoreThanOne())
	}
	return items[0], nil
}
