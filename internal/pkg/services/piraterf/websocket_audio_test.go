package piraterf

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/common-go/env"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
}

func TestSendAudioPlaylistCreateSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendAudioPlaylistCreateSuccessEvent("test_playlist", "/test/path.wav")
	})
}

func TestSendAudioPlaylistCreateErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
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
			name:        "valid request",
			eventData:   json.RawMessage(`{"files": ["/files/test_2s.mp3", "/files/test_3s.wav"]}`),
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
			hub := websocket.NewHub("test")
			defer hub.Close()

			service := &PIrateRF{
				websocketHub: hub,
				config: Config{
					FilesDir: "/workspace/.fixtures",
				},
			}
			var msg audioPlaylistCreateMessage
			err := json.Unmarshal(tt.eventData, &msg)
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