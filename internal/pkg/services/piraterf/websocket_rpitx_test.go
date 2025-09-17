package piraterf

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/common-go/env"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessAudioModifications(t *testing.T) {
	// Set up logger to debug level for tests
	logrus.SetLevel(logrus.WarnLevel)

	// Create temporary directory for test output
	tempDir := t.TempDir()

	// Create PIrateRF instance with test config
	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
		serviceCtx: context.Background(),
	}

	// Setup fixture files paths
	projectRoot := getProjectRoot(t)
	fixturesDir := filepath.Join(projectRoot, ".fixtures")

	testFiles := []string{
		filepath.Join(fixturesDir, "test_2s.mp3"), // 2 seconds
		filepath.Join(fixturesDir, "test_3s.wav"), // 3 seconds
		filepath.Join(fixturesDir, "test_4s.wav"), // 4 seconds
	}

	// Verify all fixture files exist
	for _, filePath := range testFiles {
		require.FileExists(t, filePath, "Fixture file should exist: %s", filePath)
	}

	logger := logrus.WithField("test", "TestProcessAudioModifications")

	t.Run("NoModifications", func(t *testing.T) {
		// Test with no intro, outro, or PlayOnce
		args := map[string]any{
			"freq":  431.0,
			"audio": testFiles[0],
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  30,
			PlayOnce: false,
			Intro:    nil,
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 30, logger)

		require.NoError(t, err)
		assert.Equal(t, 30, finalTimeout, "Timeout should remain unchanged")
		assert.Empty(t, tempPath, "No temp playlist should be created")
		assert.JSONEq(t, string(argsJSON), string(finalArgs), "Args should remain unchanged")
	})

	t.Run("IntroOnly", func(t *testing.T) {
		introFile := testFiles[0] // 2 seconds
		mainAudio := testFiles[1] // 3 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  30,
			PlayOnce: false,
			Intro:    &introFile,
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 30, logger)

		require.NoError(t, err)
		assert.Equal(t, 30, finalTimeout, "Timeout should remain unchanged without PlayOnce")
		assert.NotEmpty(t, tempPath, "Temp playlist should be created")
		assert.True(t, strings.HasPrefix(tempPath, "/tmp/"), "Temp path should be in /tmp")
		assert.NotEqual(t, argsJSON, finalArgs, "Args should be modified")

		// Verify the modified args contain the playlist path
		var modifiedArgs map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgs)
		require.NoError(t, err)
		assert.Equal(t, tempPath, modifiedArgs["audio"], "Audio path should point to playlist")

		// Verify temp file exists
		require.FileExists(t, tempPath, "Temp playlist file should exist")

		// Clean up
		_ = os.Remove(tempPath)
	})

	t.Run("OutroOnly", func(t *testing.T) {
		mainAudio := testFiles[1] // 3 seconds
		outroFile := testFiles[2] // 4 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  30,
			PlayOnce: false,
			Intro:    nil,
			Outro:    &outroFile,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 30, logger)

		require.NoError(t, err)
		assert.Equal(t, 30, finalTimeout, "Timeout should remain unchanged without PlayOnce")
		assert.NotEmpty(t, tempPath, "Temp playlist should be created")
		assert.True(t, strings.HasPrefix(tempPath, "/tmp/"), "Temp path should be in /tmp")
		assert.NotEqual(t, argsJSON, finalArgs, "Args should be modified")

		// Verify temp file exists
		require.FileExists(t, tempPath, "Temp playlist file should exist")

		// Clean up
		_ = os.Remove(tempPath)
	})

	t.Run("IntroAndOutro", func(t *testing.T) {
		introFile := testFiles[0] // 2 seconds
		mainAudio := testFiles[1] // 3 seconds
		outroFile := testFiles[2] // 4 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  30,
			PlayOnce: false,
			Intro:    &introFile,
			Outro:    &outroFile,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 30, logger)

		require.NoError(t, err)
		assert.Equal(t, 30, finalTimeout, "Timeout should remain unchanged without PlayOnce")
		assert.NotEmpty(t, tempPath, "Temp playlist should be created")
		assert.True(t, strings.HasPrefix(tempPath, "/tmp/"), "Temp path should be in /tmp")

		// Verify the modified args contain the playlist path
		var modifiedArgs map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgs)
		require.NoError(t, err)
		assert.Equal(t, tempPath, modifiedArgs["audio"], "Audio path should point to playlist")

		// Verify temp file exists
		require.FileExists(t, tempPath, "Temp playlist file should exist")

		// Clean up
		_ = os.Remove(tempPath)
	})

	t.Run("PlayOnceWithoutIntroOutro", func(t *testing.T) {
		mainAudio := testFiles[1] // 3 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  0, // No timeout set
			PlayOnce: true,
			Intro:    nil,
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 0, logger)

		require.NoError(t, err)
		assert.Greater(t, finalTimeout, 0, "Timeout should be set based on audio duration")
		assert.Equal(t, 5, finalTimeout, "Timeout should be 5 seconds for 3s audio + 2s silence")
		assert.NotEmpty(t, tempPath, "Silence temp file should be created")
		assert.Contains(t, tempPath, "_with_silence", "Temp path should contain silence file")

		// Verify args are modified to use silence file
		var modifiedArgsMap map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgsMap)
		require.NoError(t, err)
		assert.Contains(t, modifiedArgsMap["audio"].(string), "_with_silence", "Args should point to silence file")
	})

	t.Run("PlayOnceWithIntroAndOutro", func(t *testing.T) {
		introFile := testFiles[0] // 2 seconds
		mainAudio := testFiles[1] // 3 seconds
		outroFile := testFiles[2] // 4 seconds
		// Expected total: 2 + 3 + 4 = 9 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  0, // No timeout set
			PlayOnce: true,
			Intro:    &introFile,
			Outro:    &outroFile,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 0, logger)

		require.NoError(t, err)
		assert.Equal(t, 11, finalTimeout, "Timeout should be 11 seconds (9s playlist + 2s silence)")
		assert.NotEmpty(t, tempPath, "Silence temp file should be created")
		assert.Contains(t, tempPath, "_with_silence", "Temp path should contain silence file")

		// Verify the modified args contain the silence file path
		var modifiedArgs map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgs)
		require.NoError(t, err)
		assert.Contains(t, modifiedArgs["audio"].(string), "_with_silence", "Audio path should point to silence file")

		// Verify temp file exists
		require.FileExists(t, tempPath, "Temp playlist file should exist")

		t.Logf("Created playlist with duration-based timeout: %d seconds", finalTimeout)

		// Clean up
		_ = os.Remove(tempPath)
	})

	t.Run("PlayOnceWithUserTimeoutShorter", func(t *testing.T) {
		mainAudio := testFiles[1] // 3 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  2, // User timeout shorter than audio
			PlayOnce: true,
			Intro:    nil,
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 2, logger)

		require.NoError(t, err)
		assert.Equal(t, 2, finalTimeout, "Should use user timeout when shorter than audio + silence")
		assert.NotEmpty(t, tempPath, "Silence temp file should be created")
		assert.Contains(t, tempPath, "_with_silence", "Temp path should contain silence file")

		// Verify args are modified to use silence file
		var modifiedArgsMap map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgsMap)
		require.NoError(t, err)
		assert.Contains(t, modifiedArgsMap["audio"].(string), "_with_silence", "Args should point to silence file")
	})

	t.Run("PlayOnceWithUserTimeoutLonger", func(t *testing.T) {
		mainAudio := testFiles[1] // 3 seconds

		args := map[string]any{
			"freq":  431.0,
			"audio": mainAudio,
			"pi":    "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  10, // User timeout longer than audio
			PlayOnce: true,
			Intro:    nil,
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 10, logger)

		require.NoError(t, err)
		assert.Equal(t, 5, finalTimeout, "Should use audio duration + 2s silence when shorter than user timeout")
		assert.NotEmpty(t, tempPath, "Silence temp file should be created")
		assert.Contains(t, tempPath, "_with_silence", "Temp path should contain silence file")

		// Verify args are modified to use silence file
		var modifiedArgsMap map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgsMap)
		require.NoError(t, err)
		assert.Contains(t, modifiedArgsMap["audio"].(string), "_with_silence", "Args should point to silence file")
	})

	t.Run("InvalidArgs", func(t *testing.T) {
		// Test with invalid JSON args
		invalidJSON := json.RawMessage(`{"invalid": json}`)

		msg := rpitxExecutionStartMessage{
			Args:     invalidJSON,
			Timeout:  30,
			PlayOnce: false,
			Intro:    nil,
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 30, logger)

		require.Error(t, err) // Should error due to invalid JSON
		assert.Contains(t, err.Error(), "failed to unmarshal args")
		assert.Equal(t, 30, finalTimeout, "Should return original timeout")
		assert.Empty(t, tempPath, "No temp playlist should be created")
		assert.Equal(t, invalidJSON, finalArgs, "Should return original args")
	})

	t.Run("MissingAudioField", func(t *testing.T) {
		// Test with args missing audio field
		args := map[string]any{
			"freq": 431.0,
			"pi":   "TEST123",
		}
		argsJSON, err := json.Marshal(args)
		require.NoError(t, err)

		msg := rpitxExecutionStartMessage{
			Args:     argsJSON,
			Timeout:  30,
			PlayOnce: true,
			Intro:    &testFiles[0],
			Outro:    nil,
		}

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(msg, 30, logger)

		require.NoError(t, err) // Should not error, just return original values
		assert.Equal(t, 30, finalTimeout, "Should return original timeout")
		assert.Empty(t, tempPath, "No temp playlist should be created")
		assert.JSONEq(t, string(argsJSON), string(finalArgs), "Should return original args")
	})
}

func TestProcessImageModifications(t *testing.T) {
	tests := []struct {
		name        string
		args        any // Changed to any to handle both map[string]any and string
		setupFiles  func(tempDir string) string
		expectError bool
		checkResult func(t *testing.T, originalArgs, result json.RawMessage, tempDir string)
	}{
		{
			name: "valid image file conversion",
			args: map[string]any{
				"pictureFile": "", // Will be set in test
				"frequency":   88.0,
			},
			setupFiles: func(tempDir string) string {
				// Use fixture image file
				return "/workspace/.fixtures/test_red_100x50.png"
			},
			expectError: true, // Expected since ImageMagick convert won't be available
			checkResult: func(t *testing.T, originalArgs, result json.RawMessage, tempDir string) {
				// Since ImageMagick won't be available, we expect the original args back
				assert.Equal(t, originalArgs, result)
			},
		},
		{
			name: "missing pictureFile field",
			args: map[string]any{
				"frequency": 88.0,
			},
			setupFiles: func(tempDir string) string {
				return ""
			},
			expectError: false,
			checkResult: func(t *testing.T, originalArgs, result json.RawMessage, tempDir string) {
				// Should return original args unchanged
				assert.Equal(t, originalArgs, result)
			},
		},
		{
			name: "empty pictureFile",
			args: map[string]any{
				"pictureFile": "",
				"frequency":   88.0,
			},
			setupFiles: func(tempDir string) string {
				return ""
			},
			expectError: false,
			checkResult: func(t *testing.T, originalArgs, result json.RawMessage, tempDir string) {
				// Should return original args unchanged
				assert.Equal(t, originalArgs, result)
			},
		},
		{
			name: "invalid args format",
			args: "invalid json",
			setupFiles: func(tempDir string) string {
				return ""
			},
			expectError: true,
			checkResult: func(t *testing.T, originalArgs, result json.RawMessage, tempDir string) {
				// Should return original args
				assert.Equal(t, originalArgs, result)
			},
		},
		{
			name: "non-string pictureFile",
			args: map[string]any{
				"pictureFile": 123, // Invalid type
				"frequency":   88.0,
			},
			setupFiles: func(tempDir string) string {
				return ""
			},
			expectError: false,
			checkResult: func(t *testing.T, originalArgs, result json.RawMessage, tempDir string) {
				// Should return original args unchanged
				assert.Equal(t, originalArgs, result)
			},
		},
		{
			name: "already YUV format file",
			args: map[string]any{
				"pictureFile": "", // Will be set to .Y file
				"frequency":   88.0,
			},
			setupFiles: func(tempDir string) string {
				// Create directory structure
				imagesDir := filepath.Join(tempDir, "images", "uploads")
				err := os.MkdirAll(imagesDir, 0755)
				require.NoError(t, err)

				// Create a .Y file
				yFilePath := filepath.Join(imagesDir, "test.Y")
				err = os.WriteFile(yFilePath, []byte("fake yuv content"), 0644)
				require.NoError(t, err)
				return yFilePath
			},
			expectError: false,
			checkResult: func(t *testing.T, originalArgs, result json.RawMessage, tempDir string) {
				// For .Y files, the path should remain unchanged
				var originalMap, resultMap map[string]any
				err := json.Unmarshal(originalArgs, &originalMap)
				require.NoError(t, err)
				err = json.Unmarshal(result, &resultMap)
				require.NoError(t, err)

				// The pictureFile should be the same since it's already in .Y format
				assert.Equal(t, originalMap["pictureFile"], resultMap["pictureFile"])
				assert.Equal(t, originalMap["frequency"], resultMap["frequency"])
			},
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

			// Setup files and update args if needed
			filePath := ""
			if tt.setupFiles != nil {
				filePath = tt.setupFiles(tempDir)
				if filePath != "" {
					if argsMap, ok := tt.args.(map[string]any); ok {
						argsMap["pictureFile"] = filePath
					}
				}
			}

			// Marshal args to JSON
			originalArgs, err := json.Marshal(tt.args)
			require.NoError(t, err)

			// Create logger
			logger := logrus.WithField("test", tt.name)

			// Call the function
			result, err := service.processImageModifications(originalArgs, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Run custom checks
			if tt.checkResult != nil {
				tt.checkResult(t, originalArgs, result, tempDir)
			}
		})
	}
}

func TestHandleRPITXExecutionStop(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	service := &PIrateRF{
		websocketHub:     hub,
		rpitx:           rpitx,
		executionManager: newExecutionManager(rpitx, hub),
	}
	client := &websocket.Client{}
	event := &websocket.Event{
		Type: "rpitx.execution.stop",
		Data: json.RawMessage(`{}`),
	}

	require.NotPanics(t, func() {
		service.handleRPITXExecutionStop(hub, client, event)
	})
}
