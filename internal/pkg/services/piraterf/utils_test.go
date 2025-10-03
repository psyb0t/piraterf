package piraterf

import (
	"os"
)

// copyFile is a helper function to copy files for testing.
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0o600)
}
