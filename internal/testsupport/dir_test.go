package testsupport_test

import (
	"path/filepath"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestProjectRoot(t *testing.T) {
	root := testsupport.ProjectRoot(t)
	assert.DirExists(t, filepath.Join(root, ".git"))
}
