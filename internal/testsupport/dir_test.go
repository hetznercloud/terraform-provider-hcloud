package testsupport_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
)

func TestProjectRoot(t *testing.T) {
	root := testsupport.ProjectRoot(t)
	assert.DirExists(t, filepath.Join(root, ".git"))
}
