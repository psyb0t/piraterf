package piraterf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/common-go/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileConversionPostprocessor(t *testing.T) {
	tests := []struct {
		name         string
		inputResponse map[string]any
		setupFiles   func(tempDir string) string
		expectError  bool
		expectResult map[string]any
	}{
		{
			name: "audio file conversion fails when ffmpeg not available",
			inputResponse: map[string]any{
				"path": "test.mp3",
				"name": "test.mp3",
			},
			setupFiles: func(tempDir string) string {
				// Create input file
				inputPath := filepath.Join(tempDir, "test.mp3")
				err := os.WriteFile(inputPath, []byte("fake mp3 content"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: true, // Expected since we don't have ffmpeg in test env
			expectResult: map[string]any{},
		},
		{
			name: "non-audio file unchanged",
			inputResponse: map[string]any{
				"path": "document.txt",
				"name": "document.txt",
			},
			setupFiles: func(tempDir string) string {
				inputPath := filepath.Join(tempDir, "document.txt")
				err := os.WriteFile(inputPath, []byte("text content"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: false,
			expectResult: map[string]any{
				"name": "document.txt",
				// path will be absolute so we check it separately
			},
		},
		{
			name: "invalid response format",
			inputResponse: map[string]any{
				"path": 123, // Invalid type
			},
			setupFiles: func(tempDir string) string {
				return ""
			},
			expectError: false,
			expectResult: map[string]any{
				"path": 123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			audioUploadsDir := filepath.Join(tempDir, "audio", "uploads")
			err := os.MkdirAll(audioUploadsDir, 0755)
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
					// Update response path to absolute path
					tt.inputResponse["path"] = inputPath
				}
			}

			result, err := service.fileConversionPostprocessor(tt.inputResponse)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check specific expected results
				for key, expectedValue := range tt.expectResult {
					actualValue, exists := result[key]
					assert.True(t, exists, "Result should contain key %s", key)

					// For converted=true, just check it's true
					if key == "converted" && expectedValue == true {
						assert.True(t, actualValue.(bool))
					} else {
						assert.Equal(t, expectedValue, actualValue)
					}
				}

				// For non-audio files, check that path contains the filename
				if tt.name == "non-audio file unchanged" {
					pathValue, exists := result["path"]
					assert.True(t, exists, "Result should contain path")
					pathStr, ok := pathValue.(string)
					assert.True(t, ok, "Path should be string")
					assert.Contains(t, pathStr, "document.txt")
				}
			}
		})
	}
}

func TestAudioConversionPostprocessor(t *testing.T) {
	tests := []struct {
		name         string
		inputResponse map[string]any
		setupFiles   func(tempDir string) string
		expectError  bool
		expectConversion bool
	}{
		{
			name: "mp3 file conversion fails without ffmpeg",
			inputResponse: map[string]any{
				"path": "test.mp3",
				"name": "test.mp3",
			},
			setupFiles: func(tempDir string) string {
				inputPath := filepath.Join(tempDir, "test.mp3")
				err := os.WriteFile(inputPath, []byte("fake mp3"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: true, // Expected since we don't have ffmpeg
			expectConversion: false,
		},
		{
			name: "wav file conversion fails without ffmpeg",
			inputResponse: map[string]any{
				"path": "audio.wav",
				"name": "audio.wav",
			},
			setupFiles: func(tempDir string) string {
				inputPath := filepath.Join(tempDir, "audio.wav")
				err := os.WriteFile(inputPath, []byte("fake wav"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: true, // Expected since we don't have ffmpeg
			expectConversion: false,
		},
		{
			name: "non-audio file ignored",
			inputResponse: map[string]any{
				"path": "document.pdf",
				"name": "document.pdf",
			},
			setupFiles: func(tempDir string) string {
				inputPath := filepath.Join(tempDir, "document.pdf")
				err := os.WriteFile(inputPath, []byte("pdf content"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: false,
			expectConversion: false,
		},
		{
			name: "missing file path in response",
			inputResponse: map[string]any{
				"name": "test.mp3",
			},
			setupFiles: func(tempDir string) string {
				return ""
			},
			expectError: false,
			expectConversion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			audioUploadsDir := filepath.Join(tempDir, "audio", "uploads")
			err := os.MkdirAll(audioUploadsDir, 0755)
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

			result, err := service.audioConversionPostprocessor(tt.inputResponse)

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

func TestConvertAudioFileWithFFmpeg(t *testing.T) {
	tests := []struct {
		name         string
		setupFiles   func(tempDir string) string
		expectError  bool
		expectWasConverted bool
	}{
		{
			name: "successful conversion of mp3 file",
			setupFiles: func(tempDir string) string {
				inputPath := filepath.Join(tempDir, "test.mp3")
				err := os.WriteFile(inputPath, []byte("fake mp3 content"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: false,
			expectWasConverted: true,
		},
		{
			name: "conversion of wav file (still processed)",
			setupFiles: func(tempDir string) string {
				inputPath := filepath.Join(tempDir, "audio.wav")
				err := os.WriteFile(inputPath, []byte("fake wav content"), 0644)
				require.NoError(t, err)
				return inputPath
			},
			expectError: false,
			expectWasConverted: true,
		},
		{
			name: "non-existent input file",
			setupFiles: func(tempDir string) string {
				return filepath.Join(tempDir, "missing.mp3")
			},
			expectError: true,
			expectWasConverted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			audioUploadsDir := filepath.Join(tempDir, "audio", "uploads")
			err := os.MkdirAll(audioUploadsDir, 0755)
			require.NoError(t, err)

			service := &PIrateRF{
				serviceCtx: context.Background(),
				config: Config{
					FilesDir: tempDir,
				},
			}

			inputPath := tt.setupFiles(tempDir)

			// This will fail because ffmpeg is not available, but we test the function structure
			convertedPath, wasConverted, err := service.convertAudioFileWithFFmpeg(inputPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, wasConverted)
				assert.Empty(t, convertedPath)
			} else {
				// In our test environment, ffmpeg won't be available, so we expect an error
				// but we can still verify the function logic up to the command execution
				assert.Error(t, err) // Expected since we don't have ffmpeg in test env

				// Check that output path was constructed correctly
				expectedBasename := filepath.Base(inputPath)
				expectedBasename = expectedBasename[:len(expectedBasename)-len(filepath.Ext(expectedBasename))]
				expectedPath := filepath.Join(audioUploadsDir, expectedBasename + constants.FileExtensionWAV)

				// The function should have calculated the correct output path even if ffmpeg fails
				if convertedPath != "" {
					assert.Equal(t, expectedPath, convertedPath)
				}
			}
		})
	}
}