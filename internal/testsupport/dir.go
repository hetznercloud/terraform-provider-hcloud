package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

// ProjectRoot starts at the current working directory and looks for the
// projects root directory, i.e. the directory containing the .git directory.
//
// ProjectRoot stops as soon as it found the .git directory, or upon reaching
// the users home directory. In the later case it fails the test.
func ProjectRoot(t *testing.T) string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home directory: %v", err)
	}
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current working directory: %v", err)
	}

	for curDir != userDir {
		st, err := os.Stat(filepath.Join(curDir, ".git"))
		if os.IsNotExist(err) || !st.IsDir() {
			curDir = filepath.Dir(curDir)
			continue
		}
		if err != nil {
			t.Fatalf("check .git directory: %v", err)
		}
		return curDir
	}
	t.Fatal("could not find project root")
	return ""
}
