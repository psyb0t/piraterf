package piraterf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/common-go/env"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendFileRenameSuccessEvent(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileRenameSuccessEvent("/old/path.txt", "newname.txt")
	})
}

func TestSendFileRenameErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileRenameErrorEvent(
			"/old/path.txt",
			"newname.txt",
			"error_type",
			"test error",
		)
	})
}

func TestSendFileDeleteSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileDeleteSuccessEvent("/deleted/file.txt")
	})
}

func TestSendFileDeleteErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileDeleteErrorEvent(
			"/deleted/file.txt",
			"error_type",
			"test error",
		)
	})
}

func TestValidateFileRenameRequest(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	tempDir := t.TempDir()

	// Create test file for this test
	testFile := filepath.Join(tempDir, "test_2s.mp3")
	err := os.WriteFile(testFile, []byte("fake mp3 content"), 0o600)
	require.NoError(t, err)

	service := &PIrateRF{
		websocketHub: hub,
		config: Config{
			FilesDir: tempDir,
		},
	}

	msg := fileRenameMessage{
		FilePath: testFile,
		NewName:  "newname.mp3",
	}

	require.NotPanics(t, func() {
		_, _, ok := service.validateFileRenameRequest(msg)
		assert.True(t, ok) // Basic validation should pass
	})
}

func TestHandleFileRename(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name      string
		eventData string
		setupFile bool
		fileName  string
	}{
		{
			name:      "invalid JSON data",
			eventData: `{invalid json`,
			setupFile: false,
		},
		{
			name:      "file not found",
			eventData: `{"filePath": "/nonexistent/file.txt", "newName": "renamed.txt"}`,
			setupFile: false,
		},
		{
			name:      "valid rename request",
			eventData: `{"filePath": "test_file.txt", "newName": "renamed_file.txt"}`,
			setupFile: true,
			fileName:  "test_file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := wshub.NewHub("test")
			defer hub.Close()

			tempDir := t.TempDir()

			service := &PIrateRF{
				websocketHub: hub,
				config: Config{
					FilesDir: tempDir,
				},
			}

			// Setup test file if needed
			var originalFilePath string
			if tt.setupFile && tt.fileName != "" {
				originalFilePath = filepath.Join(tempDir, tt.fileName)
				err := os.WriteFile(
					originalFilePath,
					[]byte("test content"),
					0o600,
				)
				require.NoError(t, err)

				// Update event data with actual file path
				tt.eventData = fmt.Sprintf(
					`{"filePath": "%s", "newName": "renamed_file.txt"}`,
					originalFilePath,
				)
			}

			event := &dabluveees.Event{
				Type: "fileRename",
				ID:   uuid.New(),
				Data: json.RawMessage(tt.eventData),
			}

			client := &wshub.Client{}

			// Test that handleFileRename doesn't panic
			if tt.eventData == invalidJSONData {
				// Invalid JSON should succeed but send error event
				err := service.handleFileRename(hub, client, event)
				assert.NoError(t, err)

				return
			}

			assert.NotPanics(t, func() {
				err := service.handleFileRename(hub, client, event)
				assert.NoError(t, err)
			})

			// Clean up renamed file if rename was successful
			if tt.setupFile && tt.fileName != "" {
				renamedPath := filepath.Join(tempDir, "renamed_file.txt")
				if _, err := os.Stat(renamedPath); err == nil {
					if err := os.Remove(renamedPath); err != nil {
						t.Logf("Failed to remove renamed file: %v", err)
					}
				}
			}
		})
	}
}

func TestHandleFileDelete(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name      string
		eventData string
		setupFile bool
		fileName  string
	}{
		{
			name:      "invalid JSON data",
			eventData: `{invalid json`,
			setupFile: false,
		},
		{
			name:      "file not found",
			eventData: `{"filePath": "/nonexistent/file.txt"}`,
			setupFile: false,
		},
		{
			name:      "valid delete request",
			eventData: `{"filePath": "test_file_to_delete.txt"}`,
			setupFile: true,
			fileName:  "test_file_to_delete.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := wshub.NewHub("test")
			defer hub.Close()

			tempDir := t.TempDir()

			service := &PIrateRF{
				websocketHub: hub,
				config: Config{
					FilesDir: tempDir,
				},
			}

			// Setup test file if needed
			if tt.setupFile && tt.fileName != "" {
				filePath := filepath.Join(tempDir, tt.fileName)
				err := os.WriteFile(filePath, []byte("test content"), 0o600)
				require.NoError(t, err)

				// Update event data with actual file path
				tt.eventData = fmt.Sprintf(`{"filePath": "%s"}`, filePath)
			}

			event := &dabluveees.Event{
				Type: "fileDelete",
				ID:   uuid.New(),
				Data: json.RawMessage(tt.eventData),
			}

			client := &wshub.Client{}

			// Test that handleFileDelete doesn't panic
			if tt.eventData == invalidJSONData {
				// Invalid JSON should succeed but send error event
				err := service.handleFileDelete(hub, client, event)
				assert.NoError(t, err)

				return
			}

			assert.NotPanics(t, func() {
				err := service.handleFileDelete(hub, client, event)
				assert.NoError(t, err)
			})
		})
	}
}
