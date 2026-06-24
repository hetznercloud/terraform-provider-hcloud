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
	resourceName    string
	parentResources [][2]string
	params          url.Values
	usingName       string
	usingValue      any
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

// WithParentResource appends a parent resource to the details.
func WithParentResource(name string, value any) DetailsOption {
	return func(o *Details) {
		o.parentResources = append(o.parentResources, [2]string{name, fmt.Sprint(value)})
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
	if d.usingName == "" {
		return
	}

	fmt.Fprintf(w, " using %s", d.usingName)
	if d.usingValue == nil {
		fmt.Fprintf(w, ".")
	} else {
		fmt.Fprintf(w, ": %v", d.usingValue)
	}
}

func (d *Details) writeParentResources(w io.Writer) {
	if len(d.parentResources) == 0 {
		return
	}

	fmt.Fprint(w, "\nParent resources: ")

	for i := range d.parentResources {
		fmt.Fprintf(w, "%s=%s", d.parentResources[i][0], d.parentResources[i][1])
		if len(d.parentResources) > i+1 {
			fmt.Fprintf(w, " ")
		}
	}
}

func (d *Details) writeParams(w io.Writer) {
	if len(d.params) == 0 {
		return
	}

	fmt.Fprint(w, "\nQuery parameters: ")
	for i, key := range slices.Sorted(maps.Keys(d.params)) {
		if len(d.params[key]) == 1 {
			fmt.Fprintf(w, "%s=%s", key, d.params[key][0])
		} else {
			fmt.Fprintf(w, "%s=%v", key, d.params[key])
		}
		if len(d.params) > i+1 {
			fmt.Fprintf(w, " ")
		}
	}
}

func (d *Details) NotFound() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Resource (%s) was not found", d.resourceName)
	d.writeUsing(b)
	fmt.Fprint(b, "\n")

	d.writeParentResources(b)
	d.writeParams(b)
	return b.String()
}

func (d *Details) MoreThanOne() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Found more than one resource (%s)", d.resourceName)
	d.writeUsing(b)
	fmt.Fprint(b, "\n")

	d.writeParentResources(b)
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
