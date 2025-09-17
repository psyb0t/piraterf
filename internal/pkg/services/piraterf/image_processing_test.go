package piraterf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string)
		expectError bool
		validate    func(t *testing.T, sourcePath, destPath string)
	}{
		{
			name: "successful file move",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				sourceFile := filepath.Join(tempDir, "source.txt")
				destFile := filepath.Join(tempDir, "dest.txt")

				// Create source file
				err := os.WriteFile(sourceFile, []byte("test content"), 0644)
				require.NoError(t, err)

				return sourceFile, destFile
			},
			expectError: false,
			validate: func(t *testing.T, sourcePath, destPath string) {
				// Source should not exist
				_, err := os.Stat(sourcePath)
				assert.True(t, os.IsNotExist(err))

				// Destination should exist with same content
				content, err := os.ReadFile(destPath)
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			},
		},
		{
			name: "move non-existent file",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				sourceFile := filepath.Join(tempDir, "nonexistent.txt")
				destFile := filepath.Join(tempDir, "dest.txt")

				return sourceFile, destFile
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePath, destPath := tt.setupFunc(t)

			err := moveFile(sourcePath, destPath)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, sourcePath, destPath)
			}
		})
	}
}

func TestCopyFileStream(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string)
		expectError bool
		validate    func(t *testing.T, sourcePath, destPath string)
	}{
		{
			name: "successful file copy",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				sourceFile := filepath.Join(tempDir, "source.txt")
				destFile := filepath.Join(tempDir, "dest.txt")

				// Create source file
				err := os.WriteFile(sourceFile, []byte("test content for copy"), 0644)
				require.NoError(t, err)

				return sourceFile, destFile
			},
			expectError: false,
			validate: func(t *testing.T, sourcePath, destPath string) {
				// Both files should exist
				_, err := os.Stat(sourcePath)
				assert.NoError(t, err)

				content, err := os.ReadFile(destPath)
				assert.NoError(t, err)
				assert.Equal(t, "test content for copy", string(content))
			},
		},
		{
			name: "copy non-existent file",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				sourceFile := filepath.Join(tempDir, "nonexistent.txt")
				destFile := filepath.Join(tempDir, "dest.txt")

				return sourceFile, destFile
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePath, destPath := tt.setupFunc(t)

			err := copyFileStream(sourcePath, destPath)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, sourcePath, destPath)
			}
		})
	}
}