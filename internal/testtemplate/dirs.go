package testtemplate

import (
	"path/filepath"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
)

// ResourceTemplateDir returns the path to the directory where the templates
// for individual resources are kept.
//
// Inside the resource template directory exists a template file with a name of
// the format <resource_name>.tf.tmpl for each resource provided by the Hetzner
// Cloud Terraform provider.
//
// The individual templates must define a valid HCL snippet for the resource.
// They must not reference other resources or data sources and they must not
// import other templates.
func ResourceTemplateDir(t *testing.T) string {
	return filepath.Join(testsupport.ProjectRoot(t), "internal", "testdata", "r")
}

// DataSourceTemplateDir returns the path to the directory where the templates
// for individual data sources are kept.
//
// Inside the data source template directory exists a template file with a name
// of the format <data_source_name>.tf.tmpl for each data source provided by
// the Hetzner Cloud Terraform provider.
//
// The individual templates must define a valid HCL snippet for the data
// source. They must not reference other resources or data sources and they
// must not import other templates.
func DataSourceTemplateDir(t *testing.T) string {
	return filepath.Join(testsupport.ProjectRoot(t), "internal", "testdata", "d")
}
