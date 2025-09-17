package piraterf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  func(tempDir string) (source, dest string)
		expectError bool
		validate    func(t *testing.T, source, dest string)
	}{
		{
			name: "successful file move",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				dest = filepath.Join(tempDir, "dest.txt")

				err := os.WriteFile(source, []byte("test content"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
			validate: func(t *testing.T, source, dest string) {
				// Source should not exist
				_, err := os.Stat(source)
				assert.True(t, os.IsNotExist(err), "Source file should be deleted")

				// Destination should exist with same content
				content, err := os.ReadFile(dest)
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			},
		},
		{
			name: "move to different directory",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				destDir := filepath.Join(tempDir, "subdir")
				err := os.MkdirAll(destDir, 0755)
				require.NoError(t, err)
				dest = filepath.Join(destDir, "moved.txt")

				err = os.WriteFile(source, []byte("move me"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
			validate: func(t *testing.T, source, dest string) {
				// Verify file was moved
				_, err := os.Stat(source)
				assert.True(t, os.IsNotExist(err))

				content, err := os.ReadFile(dest)
				assert.NoError(t, err)
				assert.Equal(t, "move me", string(content))
			},
		},
		{
			name: "move non-existent file",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "nonexistent.txt")
				dest = filepath.Join(tempDir, "dest.txt")
				return source, dest
			},
			expectError: true,
			validate: func(t *testing.T, source, dest string) {
				// Destination should not exist
				_, err := os.Stat(dest)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "move to existing file (overwrite)",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				dest = filepath.Join(tempDir, "existing.txt")

				err := os.WriteFile(source, []byte("new content"), 0644)
				require.NoError(t, err)

				err = os.WriteFile(dest, []byte("old content"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
			validate: func(t *testing.T, source, dest string) {
				// Source should not exist
				_, err := os.Stat(source)
				assert.True(t, os.IsNotExist(err))

				// Destination should have new content
				content, err := os.ReadFile(dest)
				assert.NoError(t, err)
				assert.Equal(t, "new content", string(content))
			},
		},
		{
			name: "move to directory without write permissions",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				restrictedDir := filepath.Join(tempDir, "restricted")
				err := os.MkdirAll(restrictedDir, 0555) // Read-only directory
				require.NoError(t, err)
				dest = filepath.Join(restrictedDir, "dest.txt")

				err = os.WriteFile(source, []byte("test"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: true,
			validate: func(t *testing.T, source, dest string) {
				// Source should still exist
				_, err := os.Stat(source)
				assert.NoError(t, err)

				// Destination should not exist
				_, err = os.Stat(dest)
				assert.True(t, os.IsNotExist(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			source, dest := tt.setupFiles(tempDir)

			err := moveFile(source, dest)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, source, dest)
			}
		})
	}
}

func TestImageConversionPostprocessor(t *testing.T) {
	tests := []struct {
		name         string
		inputResponse map[string]any
		setupFiles   func(tempDir string) string
		expectError  bool
		expectConversion bool
	}{
		{
			name: "png image conversion succeeds with convert",
			inputResponse: map[string]any{
				"path": "/workspace/.fixtures/test_red_100x50.png",
				"name": "test_red_100x50.png",
			},
			setupFiles: func(tempDir string) string {
				return "/workspace/.fixtures/test_red_100x50.png"
			},
			expectError: false, // ImageMagick convert is available
			expectConversion: true,
		},
		{
			name: "jpg image conversion succeeds with convert",
			inputResponse: map[string]any{
				"path": "/workspace/.fixtures/test_gradient_200x100.jpg",
				"name": "test_gradient_200x100.jpg",
			},
			setupFiles: func(tempDir string) string {
				return "/workspace/.fixtures/test_gradient_200x100.jpg"
			},
			expectError: false, // ImageMagick convert is available
			expectConversion: true,
		},
		{
			name: "non-image file ignored",
			inputResponse: map[string]any{
				"path": "/workspace/.fixtures/test_document.txt",
				"name": "test_document.txt",
			},
			setupFiles: func(tempDir string) string {
				return "/workspace/.fixtures/test_document.txt"
			},
			expectError: false,
			expectConversion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			imagesUploadsDir := filepath.Join(tempDir, "images", "uploads")
			err := os.MkdirAll(imagesUploadsDir, 0755)
			require.NoError(t, err)

			service := &PIrateRF{
				serviceCtx: context.Background(),
				config: Config{
					FilesDir: tempDir,
				},
			}

			// Set up files if needed
			if tt.setupFiles != nil {
				inputPath := tt.setupFiles(tempDir)
				if inputPath != "" {
					tt.inputResponse["path"] = inputPath
				}
			}

			result, err := service.imageConversionPostprocessor(tt.inputResponse)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check if conversion happened as expected
				converted, hasConverted := result["converted"].(bool)
				if tt.expectConversion {
					assert.True(t, hasConverted, "Result should have 'converted' key")
					assert.True(t, converted, "File should be marked as converted")
				} else {
					// Either no converted key or converted=false
					if hasConverted {
						assert.False(t, converted, "File should not be marked as converted")
					}
				}
			}
		})
	}
}

func TestCopyFileStream(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  func(tempDir string) (source, dest string)
		expectError bool
		validate    func(t *testing.T, source, dest string)
	}{
		{
			name: "successful file copy",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				dest = filepath.Join(tempDir, "dest.txt")

				err := os.WriteFile(source, []byte("copy this content"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
			validate: func(t *testing.T, source, dest string) {
				// Both files should exist
				_, err := os.Stat(source)
				assert.NoError(t, err)

				content, err := os.ReadFile(dest)
				assert.NoError(t, err)
				assert.Equal(t, "copy this content", string(content))

				// Source should still exist with same content
				sourceContent, err := os.ReadFile(source)
				assert.NoError(t, err)
				assert.Equal(t, "copy this content", string(sourceContent))
			},
		},
		{
			name: "copy large file",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "large.txt")
				dest = filepath.Join(tempDir, "large_copy.txt")

				// Create a larger file to test streaming
				largeContent := make([]byte, 10240) // 10KB
				for i := range largeContent {
					largeContent[i] = byte(i % 256)
				}

				err := os.WriteFile(source, largeContent, 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
			validate: func(t *testing.T, source, dest string) {
				sourceContent, err := os.ReadFile(source)
				assert.NoError(t, err)

				destContent, err := os.ReadFile(dest)
				assert.NoError(t, err)

				assert.Equal(t, sourceContent, destContent)
				assert.Equal(t, 10240, len(destContent))
			},
		},
		{
			name: "copy non-existent file",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "nonexistent.txt")
				dest = filepath.Join(tempDir, "dest.txt")
				return source, dest
			},
			expectError: true,
			validate: func(t *testing.T, source, dest string) {
				// Destination should not exist
				_, err := os.Stat(dest)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "copy to existing file (overwrite)",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				dest = filepath.Join(tempDir, "existing.txt")

				err := os.WriteFile(source, []byte("new content"), 0644)
				require.NoError(t, err)

				err = os.WriteFile(dest, []byte("old content"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
			validate: func(t *testing.T, source, dest string) {
				// Both should exist
				_, err := os.Stat(source)
				assert.NoError(t, err)

				// Destination should have new content
				content, err := os.ReadFile(dest)
				assert.NoError(t, err)
				assert.Equal(t, "new content", string(content))
			},
		},
		{
			name: "copy to directory without write permissions",
			setupFiles: func(tempDir string) (source, dest string) {
				source = filepath.Join(tempDir, "source.txt")
				restrictedDir := filepath.Join(tempDir, "restricted")
				err := os.MkdirAll(restrictedDir, 0555) // Read-only directory
				require.NoError(t, err)
				dest = filepath.Join(restrictedDir, "dest.txt")

				err = os.WriteFile(source, []byte("test"), 0644)
				require.NoError(t, err)

				return source, dest
			},
			expectError: true,
			validate: func(t *testing.T, source, dest string) {
				// Source should still exist
				_, err := os.Stat(source)
				assert.NoError(t, err)

				// Destination should not exist
				_, err = os.Stat(dest)
				assert.True(t, os.IsNotExist(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			source, dest := tt.setupFiles(tempDir)

			err := copyFileStream(source, dest)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, source, dest)
			}
		})
	}
}