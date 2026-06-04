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
		wantItem *string
		wantDiag diag.Diagnostic
	}{
		{
			name:     "zero item",
			items:    []*string{},
			wantItem: nil,
			wantDiag: diag.NewErrorDiagnostic(
				"Resource not found",
				"Resource (item) was not found using label selector: key=value\n\nsort=id:asc\nstatus=running\n",
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
				"Found more than one resource (item) using label selector: key=value\n\nsort=id:asc\nstatus=running\n",
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
				WithResourceName("item"),
				WithUsing("label selector", "key=value"),
				WithListOpts(opts),
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
				"Resource (item) was not found using label selector: key=value\n\nsort=id:asc\nstatus=running\n",
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
