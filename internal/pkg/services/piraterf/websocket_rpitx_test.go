package piraterf

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessAudioModifications(t *testing.T) {
	// Set up logger to debug level for tests
	logrus.SetLevel(logrus.DebugLevel)

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
