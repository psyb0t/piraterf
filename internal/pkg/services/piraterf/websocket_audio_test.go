package piraterf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/env"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const invalidJSONData = `{invalid json`

func TestSendAudioPlaylistCreateSuccessEvent(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendAudioPlaylistCreateSuccessEvent("test_playlist", "/test/path.wav")
	})
}

func TestSendAudioPlaylistCreateErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendAudioPlaylistCreateErrorEvent("clientID", "eventID", "test error")
	})
}

func TestValidatePlaylistRequest(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name        string
		eventData   json.RawMessage
		expectError bool
	}{
		{
			name: "valid request",
			eventData: json.RawMessage(
				`{"files": ["/files/test_2s.mp3", "/files/test_3s.wav"]}`,
			),
			expectError: false,
		},
		{
			name:        "missing files",
			eventData:   json.RawMessage(`{}`),
			expectError: true,
		},
		{
			name:        "invalid JSON",
			eventData:   json.RawMessage(`{invalid`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := wshub.NewHub("test")
			defer hub.Close()

			tempDir := t.TempDir()

			// Create test files for this test
			testFile1 := filepath.Join(tempDir, "test_2s.mp3")
			err := os.WriteFile(testFile1, []byte("fake mp3 content"), 0o600)
			require.NoError(t, err)

			testFile2 := filepath.Join(tempDir, "test_3s.wav")
			err = os.WriteFile(testFile2, []byte("fake wav content"), 0o600)
			require.NoError(t, err)

			service := &PIrateRF{
				websocketHub: hub,
				config: Config{
					FilesDir: tempDir,
				},
			}

			var msg audioPlaylistCreateMessage

			err = json.Unmarshal(tt.eventData, &msg)
			if err != nil && tt.expectError {
				assert.Error(t, err)

				return
			}

			_, ok := service.validatePlaylistRequest(msg)

			if tt.expectError {
				assert.False(t, ok)
			} else {
				assert.True(t, ok)
			}
		})
	}
}

func TestConvertHTTPPathToFileSystem(t *testing.T) {
	tempDir := t.TempDir()
	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
	}

	tests := []struct {
		name     string
		httpPath string
		expected string
	}{
		{
			name:     "audio file path",
			httpPath: "/files/audio/test.mp3",
			expected: filepath.Join(tempDir, "audio", "test.mp3"),
		},
		{
			name:     "nested path",
			httpPath: "/files/audio/uploads/song.wav",
			expected: filepath.Join(tempDir, "audio", "uploads", "song.wav"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.convertHTTPPathToFileSystem(tt.httpPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleAudioPlaylistCreate(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name      string
		eventData string
	}{
		{
			name:      "invalid JSON data",
			eventData: invalidJSONData,
		},
		{
			name:      "missing files in request",
			eventData: `{"playlistFileName": "test.wav"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := wshub.NewHub("test")
			defer hub.Close()

			mockCmd := commander.NewMock()

			tempDir := t.TempDir()

			service := &PIrateRF{
				websocketHub: hub,
				commander:    mockCmd,
				config: Config{
					FilesDir: tempDir,
				},
			}

			event := &dabluveees.Event{
				Type: "audioPlaylistCreate",
				ID:   uuid.New(),
				Data: json.RawMessage(tt.eventData),
			}

			client := &wshub.Client{}

			// Test that handleAudioPlaylistCreate doesn't panic
			if tt.eventData == invalidJSONData {
				// Invalid JSON should succeed but send error event
				err := service.handleAudioPlaylistCreate(hub, client, event)
				assert.NoError(t, err)
			} else {
				assert.NotPanics(t, func() {
					err := service.handleAudioPlaylistCreate(hub, client, event)
					assert.NoError(t, err)
				})
			}
		})
	}
}
