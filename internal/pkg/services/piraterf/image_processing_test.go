package piraterf

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/env"
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFileStream(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupSrc    func() string
		setupDst    func() string
		expectError bool
		errorCheck  func(err error) bool
	}{
		{
			name: "source file doesn't exist",
			setupSrc: func() string {
				return "/nonexistent/source/file.txt"
			},
			setupDst: func() string {
				return filepath.Join(tempDir, "dest.txt")
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil && err.Error() != ""
			},
		},
		{
			name: "destination directory doesn't exist",
			setupSrc: func() string {
				srcFile := filepath.Join(tempDir, "source.txt")
				if err := os.WriteFile(srcFile, []byte("test content"), 0o600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}

				return srcFile
			},
			setupDst: func() string {
				return "/foo/does/not/exist/fucking/shit/file.fuck"
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil && err.Error() != ""
			},
		},
		{
			name: "successful copy",
			setupSrc: func() string {
				srcFile := filepath.Join(tempDir, "source_success.txt")
				content := []byte("test content for success")
				if err := os.WriteFile(srcFile, content, 0o600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}

				return srcFile
			},
			setupDst: func() string {
				return filepath.Join(tempDir, "dest_success.txt")
			},
			expectError: false,
			errorCheck: func(err error) bool {
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := tt.setupSrc()
			dst := tt.setupDst()

			err := copyFileStream(src, dst)

			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, tt.errorCheck(err), "Error check failed")

				return
			}

			assert.NoError(t, err)
			// Verify file was actually copied
			if err == nil {
				content, readErr := os.ReadFile(dst)
				assert.NoError(t, readErr)
				assert.NotEmpty(t, content)
			}

			assert.True(t, tt.errorCheck(err), "Error check failed")
		})
	}
}

func TestMoveFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  func(tempDir string) (source, dest string)
		expectError bool
	}{
		{
			name: "successful file move",
			setupFiles: func(tempDir string) (string, string) {
				source := filepath.Join(tempDir, "source.txt")
				dest := filepath.Join(tempDir, "dest.txt")
				err := os.WriteFile(source, []byte("test content"), 0o600)
				require.NoError(t, err)

				return source, dest
			},
			expectError: false,
		},
		{
			name: "source file does not exist",
			setupFiles: func(tempDir string) (string, string) {
				source := filepath.Join(tempDir, "nonexistent.txt")
				dest := filepath.Join(tempDir, "dest.txt")

				return source, dest
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			source, dest := tt.setupFiles(tempDir)

			err := moveFile(source, dest)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			// Verify file was moved
			assert.NoFileExists(t, source)
			assert.FileExists(t, dest)
		})
	}
}

func TestConvertImageToYUV(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name        string
		inputFile   string
		expectError bool
		mockError   bool
	}{
		{
			name:        "successful conversion",
			inputFile:   ".fixtures/test_red_100x50.png",
			expectError: false,
			mockError:   false,
		},
		{
			name:        "imagemagick command fails",
			inputFile:   ".fixtures/test_red_100x50.png",
			expectError: true,
			mockError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create required directory structure
			imagesUploadsDir := filepath.Join(tempDir, "images", "uploads")
			err := os.MkdirAll(imagesUploadsDir, 0o750)
			require.NoError(t, err)

			var mockCmd commander.Commander

			if tt.mockError {
				mock := commander.NewMock()
				mock.Expect("convert").ReturnError(ctxerrors.New("mock imagemagick error"))
				mockCmd = mock
			} else {
				// Create a custom mock commander that creates output files
				mock := &fileCreatingMockCommander{
					MockCommander: *commander.NewMock(),
				}

				// Set up mock for convert with exact argument count
				// Expected args: inputPath, "-resize", "320x", "-flip",
				// "-quantize", "YUV", "-dither", "FloydSteinberg",
				// "-colors", "4", "-interlace", "partition", outputPath
				mock.ExpectWithMatchers("convert",
					commander.Any(),                   // input path
					commander.Exact("-resize"),        // -resize
					commander.Exact("320x"),           // width
					commander.Exact("-flip"),          // flip
					commander.Exact("-quantize"),      // quantize
					commander.Exact("YUV"),            // YUV colorspace
					commander.Exact("-dither"),        // dither
					commander.Exact("FloydSteinberg"), // dithering method
					commander.Exact("-colors"),        // colors
					commander.Exact("4"),              // color count
					commander.Exact("-interlace"),     // interlace
					commander.Exact("partition"),      // partition method
					commander.Any(),                   // output path
				).ReturnOutput([]byte("mock convert output"))
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

			logger := logrus.WithField("test", "convertImage")
			outputPath, err := service.convertImageToYUV(tt.inputFile, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, outputPath)

				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, outputPath)
			assert.Contains(t, outputPath, ".Y")
		})
	}
}

func TestImageConversionPostprocessor(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name          string
		inputResponse map[string]any
		setupFiles    func(tempDir string) string
		expectError   bool
		expectResult  map[string]any
	}{
		{
			name: "image file conversion success",
			inputResponse: map[string]any{
				"path": ".fixtures/test_red_100x50.png",
				"name": "test_red_100x50.png",
			},
			setupFiles: func(tempDir string) string {
				// Create PNG file in temp directory for testing
				pngContent := []byte{
					0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
					0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
					0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x32,
					0x08, 0x02, 0x00, 0x00, 0x00, 0x91, 0x5C, 0x8F,
					0x96, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
					0x54, 0x78, 0xDA, 0x63, 0xF8, 0x0F, 0x00, 0x00,
					0x01, 0x00, 0x01, 0x00, 0x18, 0xDD, 0x8D, 0xB4,
					0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
					0xAE, 0x42, 0x60, 0x82,
				}
				testFile := filepath.Join(tempDir, "test_red_100x50.png")
				if err := os.WriteFile(testFile, pngContent, 0o600); err != nil {
					return ""
				}

				return testFile
			},
			expectError: false,
			expectResult: map[string]any{
				"converted": true,
			},
		},
		{
			name: "non-image file unchanged",
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
			imagesUploadsDir := filepath.Join(tempDir, "images", "uploads")
			err := os.MkdirAll(imagesUploadsDir, 0o750)
			require.NoError(t, err)

			// Create a custom mock commander that creates output files
			mockCmd := &fileCreatingMockCommander{
				MockCommander: *commander.NewMock(),
			}

			// Set up mock for convert with exact argument count
			// and patterns (YUV conversion)
			mockCmd.ExpectWithMatchers("convert",
				commander.Any(),                   // input path
				commander.Exact("-resize"),        // -resize
				commander.Exact("320x"),           // width
				commander.Exact("-flip"),          // flip
				commander.Exact("-quantize"),      // quantize
				commander.Exact("YUV"),            // YUV colorspace
				commander.Exact("-dither"),        // dither
				commander.Exact("FloydSteinberg"), // dithering method
				commander.Exact("-colors"),        // colors
				commander.Exact("4"),              // color count
				commander.Exact("-interlace"),     // interlace
				commander.Exact("partition"),      // partition method
				commander.Any(),                   // output path
			).ReturnOutput([]byte("mock convert YUV output"))

			// Set up mock for convert RGB conversion
			mockCmd.ExpectWithMatchers("convert",
				commander.Any(),             // input path
				commander.Exact("-resize"),  // -resize
				commander.Exact("320x256!"), // exact size with exclamation
				commander.Exact("-depth"),   // depth
				commander.Exact("8"),        // 8-bit
				commander.Any(),             // output path (rgb:path)
			).ReturnOutput([]byte("mock convert RGB output"))

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

			result, err := service.imageConversionPostprocessor(tt.inputResponse)

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

			// For non-image files, check path
			if tt.name == "non-image file unchanged" {
				pathValue, exists := result["path"]
				assert.True(t, exists, "Result should contain path")

				pathStr, ok := pathValue.(string)
				assert.True(t, ok, "Path should be string")
				assert.Contains(t, pathStr, "test_document.txt")
			}
		})
	}
}

func TestGetImageOutputPath(t *testing.T) {
	tempDir := t.TempDir()

	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
	}

	tests := []struct {
		name      string
		inputPath string
		expected  string
	}{
		{
			name:      "PNG file",
			inputPath: "/path/to/image.png",
			expected:  "image.Y",
		},
		{
			name:      "JPG file",
			inputPath: "/path/to/photo.jpg",
			expected:  "photo.Y",
		},
		{
			name:      "file without extension",
			inputPath: "/path/to/image",
			expected:  "image.Y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getImageOutputPath(tt.inputPath)
			assert.Contains(t, result, tt.expected)
			assert.Contains(t, result, tempDir)
		})
	}
}

func TestHandleYFile(t *testing.T) {
	tempDir := t.TempDir()

	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
	}

	// Create images uploads directory
	imagesDir := filepath.Join(tempDir, "images", "uploads")
	err := os.MkdirAll(imagesDir, 0o750)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupFile   func() string
		outputPath  string
		expectError bool
		expectSame  bool
	}{
		{
			name: "Y file already in correct directory",
			setupFile: func() string {
				yFile := filepath.Join(imagesDir, "test.Y")
				err := os.WriteFile(yFile, []byte("test Y content"), 0o600)
				require.NoError(t, err)

				return yFile
			},
			outputPath:  filepath.Join(imagesDir, "test.Y"),
			expectError: false,
			expectSame:  true,
		},
		{
			name: "Y file in wrong directory",
			setupFile: func() string {
				wrongDir := filepath.Join(tempDir, "wrong")
				err := os.MkdirAll(wrongDir, 0o750)
				require.NoError(t, err)
				yFile := filepath.Join(wrongDir, "test.Y")
				err = os.WriteFile(yFile, []byte("test Y content"), 0o600)
				require.NoError(t, err)

				return yFile
			},
			outputPath:  filepath.Join(imagesDir, "moved.Y"),
			expectError: false,
			expectSame:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := tt.setupFile()
			result, err := service.handleYFile(inputPath, tt.outputPath)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			if tt.expectSame {
				assert.Equal(t, inputPath, result)

				return
			}

			assert.Equal(t, tt.outputPath, result)
			// Verify file was moved
			assert.FileExists(t, tt.outputPath)
		})
	}
}

func TestProcessImageModifications(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tempDir := t.TempDir()

	// Create test mock commander that handles convert commands
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

	tests := []struct {
		name         string
		args         map[string]any
		expectError  bool
		expectChange bool
	}{
		{
			name: "valid image file",
			args: map[string]any{
				"pictureFile": ".fixtures/test_red_100x50.png",
				"frequency":   "433.92",
			},
			expectError:  false,
			expectChange: true,
		},
		{
			name: "no picture file",
			args: map[string]any{
				"frequency": "433.92",
			},
			expectError:  false,
			expectChange: false,
		},
		{
			name: "empty picture file",
			args: map[string]any{
				"pictureFile": "",
				"frequency":   "433.92",
			},
			expectError:  false,
			expectChange: false,
		},
		{
			name:        "invalid JSON",
			args:        nil, // will cause JSON unmarshal to fail
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				argsJSON []byte
				err      error
			)

			if tt.args != nil {
				argsJSON, err = json.Marshal(tt.args)
				require.NoError(t, err)
			} else {
				argsJSON = []byte("invalid json")
			}

			logger := logrus.WithField("test", "processImageModifications")
			result, err := service.processImageModifications(argsJSON, logger)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			if tt.expectChange {
				// Result should be different from input if image was converted
				assert.NotEqual(t, string(argsJSON), string(result))

				return
			}

			// Result should be same as input if no changes
			assert.Equal(t, string(argsJSON), string(result))
		})
	}
}
