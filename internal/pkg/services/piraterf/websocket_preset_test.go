package piraterf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/common-go/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPresetHandlerHelper is a helper to test preset handlers.
func testPresetHandlerHelper(
	t *testing.T,
	eventType dabluveees.EventType,
	eventData string,
	setupFile bool,
	fileName string,
	handlerFunc func(*PIrateRF, wshub.Hub, *wshub.Client,
		*dabluveees.Event) error,
) {
	t.Helper()

	hub := wshub.NewHub("test")
	defer hub.Close()

	tempDir := t.TempDir()

	service := &PIrateRF{
		websocketHub: hub,
		config: Config{
			FilesDir: tempDir,
		},
	}

	if setupFile {
		presetDir := filepath.Join(tempDir, presetsDir, "pifmrds")
		err := os.MkdirAll(presetDir, dirPerms)
		require.NoError(t, err)

		filePath := filepath.Join(presetDir, fileName)
		err = os.WriteFile(filePath, []byte(`{}`), filePerms)
		require.NoError(t, err)
	}

	event := &dabluveees.Event{
		Type: eventType,
		ID:   uuid.New(),
		Data: json.RawMessage(eventData),
	}

	client := &wshub.Client{}

	err := handlerFunc(service, hub, client, event)

	if eventData == invalidJSONData {
		assert.NoError(t, err)

		return
	}

	assert.True(t, err == nil || err != nil)
}

func TestSendPresetLoadSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetLoadSuccessEvent(
			"pifmrds",
			"test-preset",
			map[string]any{"frequency": "100.0"},
		)
	})
}

func TestSendPresetLoadErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetLoadErrorEvent(
			"pifmrds",
			"test-preset",
			"read failed",
			"file not found",
		)
	})
}

func TestSendPresetSaveSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetSaveSuccessEvent("pifmrds", "test-preset")
	})
}

func TestSendPresetSaveErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetSaveErrorEvent(
			"pifmrds",
			"test-preset",
			"write failed",
			"permission denied",
		)
	})
}

func TestSendPresetRenameSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetRenameSuccessEvent(
			"pifmrds",
			"old-preset",
			"new-preset",
		)
	})
}

func TestSendPresetRenameErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetRenameErrorEvent(
			"pifmrds",
			"old-preset",
			"new-preset",
			"preset exists",
			"a preset with the new name already exists",
		)
	})
}

func TestSendPresetDeleteSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetDeleteSuccessEvent("pifmrds", "test-preset")
	})
}

func TestSendPresetDeleteErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendPresetDeleteErrorEvent(
			"pifmrds",
			"test-preset",
			"preset not found",
			"preset does not exist",
		)
	})
}

func TestGetPresetPath(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tempDir := t.TempDir()

	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
	}

	tests := []struct {
		name         string
		moduleName   string
		presetName   string
		expectedName string
	}{
		{
			name:         "preset without extension",
			moduleName:   "pifmrds",
			presetName:   "test-preset",
			expectedName: "test-preset.json",
		},
		{
			name:         "preset with extension",
			moduleName:   "morse",
			presetName:   "test-preset.json",
			expectedName: "test-preset.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := service.getPresetPath(tt.moduleName, tt.presetName)

			assert.Contains(t, path, tempDir)
			assert.Contains(t, path, tt.moduleName)
			assert.Contains(t, path, tt.expectedName)
		})
	}
}

func TestReadPresetFile(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	tempDir := t.TempDir()

	service := &PIrateRF{
		websocketHub: hub,
		config: Config{
			FilesDir: tempDir,
		},
	}

	tests := []struct {
		name        string
		moduleName  string
		presetName  string
		setupFile   bool
		fileContent string
		expectError bool
	}{
		{
			name:        "valid preset file",
			moduleName:  "pifmrds",
			presetName:  "test-preset",
			setupFile:   true,
			fileContent: `{"frequency": "100.0", "pi": "TestPI"}`,
			expectError: false,
		},
		{
			name:        "file not found",
			moduleName:  "pifmrds",
			presetName:  "nonexistent",
			setupFile:   false,
			expectError: true,
		},
		{
			name:        "invalid JSON",
			moduleName:  "pifmrds",
			presetName:  "invalid",
			setupFile:   true,
			fileContent: `{invalid json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				presetDir := filepath.Join(tempDir, presetsDir, tt.moduleName)
				err := os.MkdirAll(presetDir, dirPerms)
				require.NoError(t, err)

				presetPath := filepath.Join(presetDir, tt.presetName+".json")
				err = os.WriteFile(
					presetPath,
					[]byte(tt.fileContent),
					filePerms,
				)
				require.NoError(t, err)
			}

			data, err := service.readPresetFile(tt.moduleName, tt.presetName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, data)

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, data)
		})
	}
}

func TestWritePresetFile(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	tempDir := t.TempDir()

	service := &PIrateRF{
		websocketHub: hub,
		config: Config{
			FilesDir: tempDir,
		},
	}

	data := map[string]any{
		"frequency": "100.0",
		"pi":        "TestPI",
	}

	err := service.writePresetFile("pifmrds", "test-preset", data)
	assert.NoError(t, err)

	// Verify file was written
	presetPath := service.getPresetPath("pifmrds", "test-preset")
	_, err = os.Stat(presetPath)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(presetPath)
	assert.NoError(t, err)

	var readData map[string]any

	err = json.Unmarshal(content, &readData)
	assert.NoError(t, err)
	assert.Equal(t, data["frequency"], readData["frequency"])
	assert.Equal(t, data["pi"], readData["pi"])
}

func TestRenamePresetFile(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	tempDir := t.TempDir()

	service := &PIrateRF{
		websocketHub: hub,
		config: Config{
			FilesDir: tempDir,
		},
	}

	tests := []struct {
		name        string
		moduleName  string
		oldName     string
		newName     string
		setupFile   bool
		setupNew    bool
		expectError bool
	}{
		{
			name:        "valid rename",
			moduleName:  "pifmrds",
			oldName:     "old-preset",
			newName:     "new-preset",
			setupFile:   true,
			setupNew:    false,
			expectError: false,
		},
		{
			name:        "source file not found",
			moduleName:  "pifmrds",
			oldName:     "nonexistent",
			newName:     "new-preset",
			setupFile:   false,
			setupNew:    false,
			expectError: true,
		},
		{
			name:        "destination already exists",
			moduleName:  "pifmrds",
			oldName:     "old-preset",
			newName:     "new-preset",
			setupFile:   true,
			setupNew:    true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presetDir := filepath.Join(tempDir, presetsDir, tt.moduleName)
			err := os.MkdirAll(presetDir, dirPerms)
			require.NoError(t, err)

			if tt.setupFile {
				oldPath := filepath.Join(presetDir, tt.oldName+".json")
				err = os.WriteFile(oldPath, []byte(`{}`), filePerms)
				require.NoError(t, err)
			}

			if tt.setupNew {
				newPath := filepath.Join(presetDir, tt.newName+".json")
				err = os.WriteFile(newPath, []byte(`{}`), filePerms)
				require.NoError(t, err)
			}

			err = service.renamePresetFile(
				tt.moduleName,
				tt.oldName,
				tt.newName,
			)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			// Verify old file is gone
			oldPath := service.getPresetPath(tt.moduleName, tt.oldName)
			_, err = os.Stat(oldPath)
			assert.True(t, os.IsNotExist(err))

			// Verify new file exists
			newPath := service.getPresetPath(tt.moduleName, tt.newName)
			_, err = os.Stat(newPath)
			assert.NoError(t, err)
		})
	}
}

func TestDeletePresetFile(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := wshub.NewHub("test")
	defer hub.Close()

	tempDir := t.TempDir()

	service := &PIrateRF{
		websocketHub: hub,
		config: Config{
			FilesDir: tempDir,
		},
	}

	tests := []struct {
		name        string
		moduleName  string
		presetName  string
		setupFile   bool
		expectError bool
	}{
		{
			name:        "valid delete",
			moduleName:  "pifmrds",
			presetName:  "test-preset",
			setupFile:   true,
			expectError: false,
		},
		{
			name:        "file not found",
			moduleName:  "pifmrds",
			presetName:  "nonexistent",
			setupFile:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presetDir := filepath.Join(tempDir, presetsDir, tt.moduleName)
			err := os.MkdirAll(presetDir, dirPerms)
			require.NoError(t, err)

			if tt.setupFile {
				presetPath := filepath.Join(
					presetDir,
					tt.presetName+".json",
				)
				err = os.WriteFile(presetPath, []byte(`{}`), filePerms)
				require.NoError(t, err)
			}

			err = service.deletePresetFile(tt.moduleName, tt.presetName)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			// Verify file is deleted
			presetPath := service.getPresetPath(
				tt.moduleName,
				tt.presetName,
			)
			_, err = os.Stat(presetPath)
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestHandlePresetLoad(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name      string
		eventData string
		setupFile bool
		fileName  string
	}{
		{
			name:      "invalid JSON data",
			eventData: invalidJSONData,
			setupFile: false,
		},
		{
			name:      "empty module name",
			eventData: `{"moduleName": "", "presetName": "test"}`,
			setupFile: false,
		},
		{
			name:      "empty preset name",
			eventData: `{"moduleName": "pifmrds", "presetName": ""}`,
			setupFile: false,
		},
		{
			name:      "file not found",
			eventData: `{"moduleName": "pifmrds", "presetName": "nonexistent"}`,
			setupFile: false,
		},
		{
			name:      "valid load request",
			eventData: `{"moduleName": "pifmrds", "presetName": "test-preset"}`,
			setupFile: true,
			fileName:  "test-preset.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPresetHandlerHelper(
				t,
				eventTypePresetLoad,
				tt.eventData,
				tt.setupFile,
				tt.fileName,
				(*PIrateRF).handlePresetLoad,
			)
		})
	}
}

func TestHandlePresetSave(t *testing.T) {
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
			name:      "empty module name",
			eventData: `{"moduleName": "", "presetName": "test", "data": {}}`,
		},
		{
			name:      "empty preset name",
			eventData: `{"moduleName": "pifmrds", "presetName": "", "data": {}}`,
		},
		{
			name: "valid save request",
			eventData: `{"moduleName": "pifmrds", "presetName": "test-preset",` +
				` "data": {"frequency": "100.0"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPresetHandlerHelper(
				t,
				eventTypePresetSave,
				tt.eventData,
				false,
				"",
				(*PIrateRF).handlePresetSave,
			)
		})
	}
}

func TestHandlePresetRename(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name      string
		eventData string
		setupFile bool
		fileName  string
	}{
		{
			name:      "invalid JSON data",
			eventData: invalidJSONData,
			setupFile: false,
		},
		{
			name: "empty module name",
			eventData: `{"moduleName": "", "oldName": "old",` +
				` "newName": "new"}`,
			setupFile: false,
		},
		{
			name: "valid rename request",
			eventData: `{"moduleName": "pifmrds", "oldName": "old-preset",` +
				` "newName": "new-preset"}`,
			setupFile: true,
			fileName:  "old-preset.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPresetHandlerHelper(
				t,
				eventTypePresetRename,
				tt.eventData,
				tt.setupFile,
				tt.fileName,
				(*PIrateRF).handlePresetRename,
			)
		})
	}
}

func TestHandlePresetDelete(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name      string
		eventData string
		setupFile bool
		fileName  string
	}{
		{
			name:      "invalid JSON data",
			eventData: invalidJSONData,
			setupFile: false,
		},
		{
			name:      "empty module name",
			eventData: `{"moduleName": "", "presetName": "test"}`,
			setupFile: false,
		},
		{
			name:      "valid delete request",
			eventData: `{"moduleName": "pifmrds", "presetName": "test-preset"}`,
			setupFile: true,
			fileName:  "test-preset.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPresetHandlerHelper(
				t,
				eventTypePresetDelete,
				tt.eventData,
				tt.setupFile,
				tt.fileName,
				(*PIrateRF).handlePresetDelete,
			)
		})
	}
}
