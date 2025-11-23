package piraterf

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/constants"
	"github.com/psyb0t/common-go/env"
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gorpitx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileConversionPostprocessor(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name          string
		inputResponse map[string]any
		setupFiles    func(tempDir string) string
		expectError   bool
		expectResult  map[string]any
	}{
		{
			name: "audio file conversion success",
			inputResponse: map[string]any{
				"path":   ".fixtures/test_2s.mp3",
				"name":   "test_2s.mp3",
				"module": "pifmrds",
			},
			setupFiles: func(_ string) string {
				return ".fixtures/test_2s.mp3"
			},
			expectError: false,
			expectResult: map[string]any{
				"converted": true,
			},
		},
		{
			name: "non-audio file unchanged",
			inputResponse: map[string]any{
				"path": ".fixtures/test_document.txt",
				"name": "test_document.txt",
			},
			setupFiles: func(_ string) string {
				return ".fixtures/test_document.txt"
			},
			expectError: false,
			expectResult: map[string]any{
				"name": "test_document.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			audioUploadsDir := filepath.Join(tempDir, "audio", "uploads")
			err := os.MkdirAll(audioUploadsDir, 0o750)
			require.NoError(t, err)

			// Create a custom mock commander that creates output files
			mockCmd := &fileCreatingMockCommander{
				MockCommander: *commander.NewMock(),
			}

			// Set up mock for ffmpeg with exact argument count
			// Expected args: "-i", inputPath, "-ar", "48000", "-ac",
			// "1", "-c:a", "pcm_s16le", "-y", outputPath
			mockCmd.ExpectWithMatchers("ffmpeg",
				commander.Exact("-i"),        // -i
				commander.Any(),              // input path
				commander.Exact("-ar"),       // -ar
				commander.Exact("48000"),     // sample rate
				commander.Exact("-ac"),       // -ac
				commander.Exact("1"),         // mono channels
				commander.Exact("-c:a"),      // -c:a
				commander.Exact("pcm_s16le"), // codec
				commander.Exact("-y"),        // overwrite
				commander.Any(),              // output path
			).ReturnOutput([]byte("mock ffmpeg output"))

			service := &PIrateRF{
				serviceCtx: context.Background(),
				config: Config{
					FilesDir: tempDir,
				},
				commander: mockCmd,
				rpitx:     gorpitx.GetInstance(),
			}

			// Set up files if needed
			if tt.setupFiles != nil {
				inputPath := tt.setupFiles(tempDir)
				if inputPath != "" {
					tt.inputResponse["path"] = inputPath
				}
			}

			// Create a mock HTTP request with module form value
			req := &http.Request{
				Form: url.Values{},
			}
			if module, ok := tt.inputResponse["module"].(string); ok {
				req.Form.Set("module", module)
			}

			result, err := service.fileConversionPostprocessor(tt.inputResponse, req)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			// Check specific expected results
			for key, expectedValue := range tt.expectResult {
				actualValue, exists := result[key]
				assert.True(t, exists, "Result should contain key %s", key)

				if key == "converted" && expectedValue == true {
					boolVal, ok := actualValue.(bool)
					if !ok {
						t.Errorf(
							"Expected bool for 'converted', got %T",
							actualValue,
						)

						continue
					}

					assert.True(t, boolVal)

					continue
				}

				assert.Equal(t, expectedValue, actualValue)
			}

			// For non-audio files, check path
			if tt.name == "non-audio file unchanged" {
				pathValue, exists := result["path"]
				assert.True(t, exists, "Result should contain path")

				pathStr, ok := pathValue.(string)
				assert.True(t, ok, "Path should be string")
				assert.Contains(t, pathStr, "test_document.txt")
			}
		})
	}
}

func TestAudioConversionPostprocessor(t *testing.T) {
	tests := []struct {
		name             string
		inputResponse    map[string]any
		setupFiles       func(tempDir string) string
		expectError      bool
		expectConversion bool
	}{
		{
			name: "mp3 file conversion",
			inputResponse: map[string]any{
				"path": ".fixtures/test_2s.mp3",
				"name": "test_2s.mp3",
			},
			setupFiles: func(_ string) string {
				return ".fixtures/test_2s.mp3"
			},
			expectError:      false,
			expectConversion: true,
		},
		{
			name: "non-audio file ignored",
			inputResponse: map[string]any{
				"path": ".fixtures/test_red_100x50.png",
				"name": "test_red_100x50.png",
			},
			setupFiles: func(_ string) string {
				return ".fixtures/test_red_100x50.png"
			},
			expectError:      false,
			expectConversion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			audioUploadsDir := filepath.Join(tempDir, "audio", "uploads")
			err := os.MkdirAll(audioUploadsDir, 0o750)
			require.NoError(t, err)

			// Create a custom mock commander that creates output files
			mockCmd := &fileCreatingMockCommander{
				MockCommander: *commander.NewMock(),
			}

			// Set up mock for ffmpeg with exact argument count and patterns
			mockCmd.ExpectWithMatchers("ffmpeg",
				commander.Exact("-i"),        // -i
				commander.Any(),              // input path
				commander.Exact("-ar"),       // -ar
				commander.Exact("48000"),     // sample rate
				commander.Exact("-ac"),       // -ac
				commander.Exact("1"),         // mono channels
				commander.Exact("-c:a"),      // -c:a
				commander.Exact("pcm_s16le"), // codec
				commander.Exact("-y"),        // overwrite
				commander.Any(),              // output path
			).ReturnOutput([]byte("mock ffmpeg output"))

			service := &PIrateRF{
				serviceCtx: context.Background(),
				config: Config{
					FilesDir: tempDir,
				},
				commander: mockCmd,
				rpitx:     gorpitx.GetInstance(),
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

				return
			}

			assert.NoError(t, err)

			// Check if conversion happened as expected
			converted, hasConverted := result["converted"].(bool)
			if tt.expectConversion {
				assert.True(t, hasConverted, "Result should have 'converted' key")
				assert.True(t, converted, "File should be marked as converted")
			} else if hasConverted {
				// Either no converted key or converted=false
				assert.False(t, converted, "File should not be marked as converted")
			}
		})
	}
}

func TestConvertAudioFileWithFFmpeg(t *testing.T) {
	tests := []struct {
		name        string
		inputFile   string
		expectError bool
		mockError   bool
	}{
		{
			name:        "successful conversion",
			inputFile:   ".fixtures/test_2s.mp3",
			expectError: false,
			mockError:   false,
		},
		{
			name:        "ffmpeg command fails",
			inputFile:   ".fixtures/test_2s.mp3",
			expectError: true,
			mockError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			audioUploadsDir := filepath.Join(tempDir, "audio", "uploads")
			err := os.MkdirAll(audioUploadsDir, 0o750)
			require.NoError(t, err)

			var mockCmd commander.Commander

			if tt.mockError {
				mock := commander.NewMock()
				mock.Expect("ffmpeg").ReturnError(ctxerrors.New("mock ffmpeg error"))
				mockCmd = mock
			} else {
				// Create a custom mock commander that creates output files
				mock := &fileCreatingMockCommander{
					MockCommander: *commander.NewMock(),
				}

				// Set up mock for ffmpeg with exact argument count and patterns
				mock.ExpectWithMatchers("ffmpeg",
					commander.Exact("-i"),        // -i
					commander.Any(),              // input path
					commander.Exact("-ar"),       // -ar
					commander.Exact("48000"),     // sample rate
					commander.Exact("-ac"),       // -ac
					commander.Exact("1"),         // mono channels
					commander.Exact("-c:a"),      // -c:a
					commander.Exact("pcm_s16le"), // codec
					commander.Exact("-y"),        // overwrite
					commander.Any(),              // output path
				).ReturnOutput([]byte("mock ffmpeg output"))
				mockCmd = mock
			}

			service := &PIrateRF{
				serviceCtx: context.Background(),
				config: Config{
					FilesDir: tempDir,
				},
				commander: mockCmd,
				rpitx:     gorpitx.GetInstance(),
			}

			convertedPath, wasConverted, err := service.
				convertAudioFileWithFFmpeg(tt.inputFile)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, wasConverted)
				assert.Empty(t, convertedPath)

				return
			}

			assert.NoError(t, err)
			assert.True(t, wasConverted)
			assert.NotEmpty(t, convertedPath)

			// Check that output path was constructed correctly
			expectedBasename := filepath.Base(tt.inputFile)
			ext := filepath.Ext(expectedBasename)
			expectedBasename = expectedBasename[:len(expectedBasename)-len(ext)]
			expectedPath := filepath.Join(
				audioUploadsDir,
				expectedBasename+constants.FileExtensionWAV,
			)
			assert.Equal(t, expectedPath, convertedPath)
		})
	}
}

func TestEnsureWavExtension(t *testing.T) {
	service := &PIrateRF{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "file without extension",
			input:    "test",
			expected: "test.wav",
		},
		{
			name:     "file with wav extension",
			input:    "test.wav",
			expected: "test.wav",
		},
		{
			name:     "file with different extension",
			input:    "test.mp3",
			expected: "test.mp3.wav",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ensureWavExtension(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPlaylistOutputPath(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tempDir := t.TempDir()

	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
		rpitx: gorpitx.GetInstance(),
	}

	tests := []struct {
		name         string
		playlistName string
		outputDir    []string
		expectPath   string
	}{
		{
			name:         "with output directory",
			playlistName: "test_playlist",
			outputDir:    []string{"/tmp"},
			expectPath:   "/tmp/test_playlist.wav",
		},
		{
			name:         "without output directory",
			playlistName: "test_playlist",
			outputDir:    []string{},
			expectPath: filepath.Join(
				tempDir,
				"audio",
				"uploads",
				"test_playlist.wav",
			),
		},
		{
			name:         "empty output directory",
			playlistName: "test_playlist",
			outputDir:    []string{""},
			expectPath: filepath.Join(
				tempDir,
				"audio",
				"uploads",
				"test_playlist.wav",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getPlaylistOutputPath(tt.playlistName, tt.outputDir...)
			if tt.expectPath == "" {
				assert.Equal(t, "", result)

				return
			}

			assert.Contains(t, result, "test_playlist.wav")
		})
	}
}

func TestCreatePlaylistFromFiles(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tempDir := t.TempDir()

	// Create test mock commander that handles sox playlist creation
	mockCommander := &testMockCommander{
		tempDir: tempDir,
	}

	service := &PIrateRF{
		serviceCtx: context.Background(),
		config: Config{
			FilesDir: tempDir,
		},
		commander: mockCommander,
		rpitx:     gorpitx.GetInstance(),
	}

	// Create test input files
	testFiles := []string{
		".fixtures/test_2s.mp3",
		".fixtures/test_3s.wav",
		".fixtures/test_4s.wav",
	}

	tests := []struct {
		name         string
		playlistName string
		filePaths    []string
		outputDir    []string
		expectError  bool
	}{
		{
			name:         "create playlist with output directory",
			playlistName: "test_playlist",
			filePaths:    testFiles,
			outputDir:    []string{tempDir},
			expectError:  false,
		},
		{
			name:         "create playlist without output directory",
			playlistName: "test_playlist2.wav",
			filePaths:    testFiles,
			outputDir:    []string{},
			expectError:  false,
		},
		{
			name:         "playlist name without wav extension",
			playlistName: "no_extension",
			filePaths:    testFiles,
			outputDir:    []string{tempDir},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				outputPath string
				err        error
			)

			if len(tt.outputDir) > 0 {
				outputPath, err = service.createPlaylistFromFiles(
					tt.playlistName,
					tt.filePaths,
					tt.outputDir[0],
				)
			} else {
				outputPath, err = service.createPlaylistFromFiles(
					tt.playlistName,
					tt.filePaths,
				)
			}

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, outputPath)
			assert.Contains(
				t,
				outputPath,
				".wav",
				"Output path should contain .wav extension",
			)
		})
	}
}
