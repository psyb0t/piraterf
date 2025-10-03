package piraterf

import (
	"context"
	"encoding/json"
	"testing"

	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/env"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	introFile     = ".fixtures/test_2s.mp3"
	mainAudioFile = ".fixtures/test_3s.wav"
	outroFile     = ".fixtures/test_4s.wav"
)

func TestProcessAudioModifications(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)

	tempDir := t.TempDir()

	// Create test mock commander that handles sox commands
	wrappedMock := &testMockCommander{
		tempDir: tempDir,
	}

	service := &PIrateRF{
		serviceCtx: context.Background(),
		config: Config{
			FilesDir: tempDir,
		},
		commander: wrappedMock,
	}

	// Create test files
	testFiles := []string{
		introFile,
		mainAudioFile,
		outroFile,
	}

	// Create logger for testing
	logger := logrus.WithField("test", "processAudioModifications")

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

		finalTimeout, tempPath, finalArgs, err := service.processAudioModifications(
			msg,
			0,
			logger,
		)

		require.NoError(t, err)
		assert.Equal(
			t,
			3,
			finalTimeout,
			"Timeout matches audio duration",
		)
		assert.NotEmpty(t, tempPath, "Temp file should be created")
		assert.Contains(
			t,
			tempPath,
			"_with_silence",
			"Temp path should contain silence file",
		)

		// Verify the modified args contain the silence file path
		var modifiedArgsMap map[string]any

		err = json.Unmarshal(finalArgs, &modifiedArgsMap)
		require.NoError(t, err)

		if audioVal, ok := modifiedArgsMap["audio"].(string); ok {
			assert.Contains(
				t,
				audioVal,
				"_with_silence",
				"Args should point to silence file",
			)
		} else {
			t.Errorf("Expected string value for 'audio' key, got %T", modifiedArgsMap["audio"])
		}
	})
}

func TestGetAudioDurationWithSox(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	mockCmd := commander.NewMock()
	mockCmd.Expect("sox", "--info", "-D", "test.wav").ReturnOutput([]byte("3.500000"))

	service := &PIrateRF{
		serviceCtx: context.Background(),
		commander:  mockCmd,
	}

	duration, err := service.getAudioDurationWithSox("test.wav")
	assert.NoError(t, err)
	assert.Equal(t, 3.5, duration)
}

func TestProcessIntroOutro(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tempDir := t.TempDir()

	// Create test mock commander that handles sox commands
	mockCommander := &testMockCommander{
		tempDir: tempDir,
	}

	service := &PIrateRF{
		serviceCtx: context.Background(),
		config: Config{
			FilesDir: tempDir,
		},
		commander: mockCommander,
	}

	intro := introFile
	mainAudio := mainAudioFile
	outro := outroFile

	msg := rpitxExecutionStartMessage{
		Intro: &intro,
		Outro: &outro,
	}
	args := map[string]any{"audio": mainAudio}
	logger := logrus.WithField("test", "processIntroOutro")

	playlistPath, finalArgs, err := service.processIntroOutro(
		msg,
		args,
		mainAudio,
		logger,
	)
	assert.NoError(t, err)
	assert.Contains(t, playlistPath, ".wav")
	assert.NotNil(t, finalArgs)
}

func TestHandleRPITXExecutionStop(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{
		websocketHub:     hub,
		executionManager: newExecutionManager(gorpitx.GetInstance(), hub),
	}

	client := &wshub.Client{}
	event := &dabluveees.Event{
		Type: "rpitx.execution.stop",
		Data: json.RawMessage(`{}`),
	}

	require.NotPanics(t, func() {
		err := service.handleRPITXExecutionStop(hub, client, event)
		assert.NoError(t, err)
	})
}

func TestProcessPlayOnceTimeoutEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	// Create test mock commander
	mockCommander := &testMockCommander{
		tempDir: tempDir,
	}

	service := &PIrateRF{
		serviceCtx: context.Background(),
		config: Config{
			FilesDir: tempDir,
		},
		commander: mockCommander,
	}

	tests := []struct {
		name            string
		playOnce        bool
		originalTimeout int
		audioFile       string
		expectError     bool
		expectedTimeout int
	}{
		{
			name:            "play once disabled",
			playOnce:        false,
			originalTimeout: 10,
			audioFile:       mainAudioFile,
			expectError:     false,
			expectedTimeout: 10, // Should return original timeout
		},
		{
			name:            "play once with timeout set",
			playOnce:        true,
			originalTimeout: 15,
			audioFile:       mainAudioFile,
			expectError:     false,
			// Should return audio duration (3) lower than timeout (15)
			expectedTimeout: 3,
		},
		{
			name:            "play once without timeout",
			playOnce:        true,
			originalTimeout: 0,
			audioFile:       mainAudioFile,
			expectError:     false,
			expectedTimeout: 3, // Should calculate based on audio duration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := rpitxExecutionStartMessage{
				PlayOnce: tt.playOnce,
				Timeout:  tt.originalTimeout,
			}

			logger := logrus.WithField("test", "processPlayOnceTimeout")
			result, err := service.processPlayOnceTimeout(
				msg,
				tt.audioFile,
				tt.originalTimeout,
				logger,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTimeout, result)
			}
		})
	}
}

func TestGenerateSilenceFilePath(t *testing.T) {
	service := &PIrateRF{}

	path1 := service.generateSilenceFilePath()
	path2 := service.generateSilenceFilePath()

	// Paths should be different (different UUIDs)
	assert.NotEqual(t, path1, path2)

	// Paths should contain the expected components
	assert.Contains(t, path1, "/tmp/")
	assert.Contains(t, path1, "_with_silence.wav")
	assert.Contains(t, path2, "/tmp/")
	assert.Contains(t, path2, "_with_silence.wav")
}

func TestUpdateArgsWithSilenceFile(t *testing.T) {
	service := &PIrateRF{}

	tests := []struct {
		name              string
		modifiedArgs      json.RawMessage
		silenceAudioPath  string
		expectError       bool
		expectedAudioPath string
	}{
		{
			name: "valid args",
			modifiedArgs: json.RawMessage(
				`{"audio": "original.wav", "freq": 433.92}`,
			),
			silenceAudioPath:  "/tmp/silence.wav",
			expectError:       false,
			expectedAudioPath: "/tmp/silence.wav",
		},
		{
			name:         "invalid JSON",
			modifiedArgs: json.RawMessage(`{invalid json`),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.updateArgsWithSilenceFile(
				tt.modifiedArgs,
				tt.silenceAudioPath,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the audio path was updated
				var resultMap map[string]any

				err = json.Unmarshal(result, &resultMap)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAudioPath, resultMap["audio"])
			}
		})
	}
}
