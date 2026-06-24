package hcloudutil

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestGetOne(t *testing.T) {
	testCases := []struct {
		name     string
		items    []*string
		opts     []DetailsOption
		wantItem *string
		wantDiag diag.Diagnostic
	}{
		{
			name:     "zero item",
			items:    []*string{},
			wantItem: nil,
			wantDiag: diag.NewErrorDiagnostic(
				"Resource not found",
				"Resource (item) was not found using label selector: key=value\n\nQuery parameters: sort=id:asc status=running",
			),
		},
		{
			name:  "zero item with one parent",
			items: []*string{},
			opts: []DetailsOption{
				WithParentResource("zone", "example.com"),
			},
			wantItem: nil,
			wantDiag: diag.NewErrorDiagnostic(
				"Resource not found",
				"Resource (item) was not found using label selector: key=value\n\nParent resources: zone=example.com\nQuery parameters: sort=id:asc status=running",
			),
		},
		{
			name:  "zero item with two parents",
			items: []*string{},
			opts: []DetailsOption{
				WithParentResource("zone", "example.com"),
				WithParentResource("zone rrset", "www/A"),
			},
			wantItem: nil,
			wantDiag: diag.NewErrorDiagnostic(
				"Resource not found",
				"Resource (item) was not found using label selector: key=value\n\nParent resources: zone=example.com zone rrset=www/A\nQuery parameters: sort=id:asc status=running",
			),
		},
		{
			name:     "one item",
			items:    []*string{new("one")},
			wantItem: new("one"),
			wantDiag: nil,
		},
		{
			name:     "two items",
			items:    []*string{new("one"), new("two")},
			wantItem: nil,
			wantDiag: diag.NewErrorDiagnostic(
				"Found more than one resource",
				"Found more than one resource (item) using label selector: key=value\n\nQuery parameters: sort=id:asc status=running",
			),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			opts := hcloud.ActionListOpts{
				Status: []hcloud.ActionStatus{hcloud.ActionStatusRunning},
				Sort:   []string{"id:asc"},
			}

			item, newDiag := GetOne(tt.items,
				append(tt.opts,
					WithResourceName("item"),
					WithUsing("label selector", "key=value"),
					WithListOpts(opts),
				)...,
			)
			require.Equal(t, tt.wantItem, item)
			require.Equal(t, tt.wantDiag, newDiag)
		})
	}
}

func TestGetFirst(t *testing.T) {
	testCases := []struct {
		name     string
		items    []*string
		wantItem *string
		wantDiag diag.Diagnostic
	}{
		{
			name:     "zero item",
			items:    []*string{},
			wantItem: nil,
			wantDiag: diag.NewErrorDiagnostic(
				"Resource not found",
				"Resource (item) was not found using label selector: key=value\n\nQuery parameters: sort=id:asc status=running",
			),
		},
		{
			name:     "one item",
			items:    []*string{new("one")},
			wantItem: new("one"),
			wantDiag: nil,
		},
		{
			name:     "two items",
			items:    []*string{new("one"), new("two")},
			wantItem: new("one"),
			wantDiag: nil,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			opts := hcloud.ActionListOpts{
				Status: []hcloud.ActionStatus{hcloud.ActionStatusRunning},
				Sort:   []string{"id:asc"},
			}

			item, newDiag := GetFirst(tt.items,
				WithResourceName("item"),
				WithUsing("label selector", "key=value"),
				WithListOpts(opts),
			)

			require.Equal(t, tt.wantItem, item)
			require.Equal(t, tt.wantDiag, newDiag)
		})
	}
}
