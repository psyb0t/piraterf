package piraterf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// getProjectRoot finds the project root by looking for go.mod file.
func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up
	dir, err := os.Getwd()
	require.NoError(t, err, "Should be able to get current directory")

	for {
		// Check if go.mod exists in current directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			t.Fatal("Could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// copyFile is a helper function to copy files for testing
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0o644)
}
