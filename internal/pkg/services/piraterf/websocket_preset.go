package piraterf

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/common-go/constants"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

const (
	eventTypePresetLoad          dabluveees.EventType = "preset.load"
	eventTypePresetLoadSuccess   dabluveees.EventType = "preset.load.success"
	eventTypePresetLoadError     dabluveees.EventType = "preset.load.error"
	eventTypePresetSave          dabluveees.EventType = "preset.save"
	eventTypePresetSaveSuccess   dabluveees.EventType = "preset.save.success"
	eventTypePresetSaveError     dabluveees.EventType = "preset.save.error"
	eventTypePresetRename        dabluveees.EventType = "preset.rename"
	eventTypePresetRenameSuccess dabluveees.EventType = "preset.rename.success"
	eventTypePresetRenameError   dabluveees.EventType = "preset.rename.error"
	eventTypePresetDelete        dabluveees.EventType = "preset.delete"
	eventTypePresetDeleteSuccess dabluveees.EventType = "preset.delete.success"
	eventTypePresetDeleteError   dabluveees.EventType = "preset.delete.error"
)

type presetLoadMessage struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
}

type presetLoadSuccessMessageData struct {
	ModuleName string         `json:"moduleName"`
	PresetName string         `json:"presetName"`
	Data       map[string]any `json:"data"`
	Timestamp  int64          `json:"timestamp"`
}

type presetLoadErrorMessageData struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	Timestamp  int64  `json:"timestamp"`
}

type presetSaveMessage struct {
	ModuleName string         `json:"moduleName"`
	PresetName string         `json:"presetName"`
	Data       map[string]any `json:"data"`
}

type presetSaveSuccessMessageData struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
	Timestamp  int64  `json:"timestamp"`
}

type presetSaveErrorMessageData struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	Timestamp  int64  `json:"timestamp"`
}

type presetRenameMessage struct {
	ModuleName string `json:"moduleName"`
	OldName    string `json:"oldName"`
	NewName    string `json:"newName"`
}

type presetRenameSuccessMessageData struct {
	ModuleName string `json:"moduleName"`
	OldName    string `json:"oldName"`
	NewName    string `json:"newName"`
	Timestamp  int64  `json:"timestamp"`
}

type presetRenameErrorMessageData struct {
	ModuleName string `json:"moduleName"`
	OldName    string `json:"oldName"`
	NewName    string `json:"newName"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	Timestamp  int64  `json:"timestamp"`
}

type presetDeleteMessage struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
}

type presetDeleteSuccessMessageData struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
	Timestamp  int64  `json:"timestamp"`
}

type presetDeleteErrorMessageData struct {
	ModuleName string `json:"moduleName"`
	PresetName string `json:"presetName"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	Timestamp  int64  `json:"timestamp"`
}

func (s *PIrateRF) handlePresetLoad(
	_ wshub.Hub,
	_ *wshub.Client,
	event *dabluveees.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("Preset load requested")

	var msg presetLoadMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		s.sendPresetLoadErrorEvent("", "", "invalid request", err.Error())

		return nil
	}

	if msg.ModuleName == "" || msg.PresetName == "" {
		s.sendPresetLoadErrorEvent(
			msg.ModuleName, msg.PresetName, "invalid request",
			"module name and preset name are required",
		)

		return nil
	}

	presetData, err := s.readPresetFile(msg.ModuleName, msg.PresetName)
	if err != nil {
		return err
	}

	s.sendPresetLoadSuccessEvent(msg.ModuleName, msg.PresetName, presetData)

	return nil
}

func (s *PIrateRF) readPresetFile(
	moduleName, presetName string,
) (map[string]any, error) {
	presetPath := s.getPresetPath(moduleName, presetName)

	data, err := os.ReadFile(presetPath)
	if err != nil {
		s.sendPresetLoadErrorEvent(
			moduleName, presetName, "read failed", err.Error(),
		)

		return nil, ctxerrors.Wrap(err, "failed to read preset file")
	}

	var presetData map[string]any
	if err := json.Unmarshal(data, &presetData); err != nil {
		s.sendPresetLoadErrorEvent(
			moduleName, presetName, "parse failed", err.Error(),
		)

		return nil, ctxerrors.Wrap(err, "failed to parse preset JSON")
	}

	return presetData, nil
}

func (s *PIrateRF) handlePresetSave(
	_ wshub.Hub,
	_ *wshub.Client,
	event *dabluveees.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("Preset save requested")

	var msg presetSaveMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		s.sendPresetSaveErrorEvent("", "", "invalid request", err.Error())

		return nil
	}

	if msg.ModuleName == "" || msg.PresetName == "" {
		s.sendPresetSaveErrorEvent(
			msg.ModuleName, msg.PresetName, "invalid request",
			"module name and preset name are required",
		)

		return nil
	}

	if err := s.writePresetFile(
		msg.ModuleName, msg.PresetName, msg.Data,
	); err != nil {
		return err
	}

	s.sendPresetSaveSuccessEvent(msg.ModuleName, msg.PresetName)

	return nil
}

func (s *PIrateRF) writePresetFile(
	moduleName, presetName string, data map[string]any,
) error {
	modulePresetDir := path.Join(s.config.FilesDir, presetsDir, moduleName)
	if err := os.MkdirAll(modulePresetDir, dirPerms); err != nil {
		s.sendPresetSaveErrorEvent(
			moduleName, presetName, "directory creation failed", err.Error(),
		)

		return ctxerrors.Wrap(err, "failed to create module preset directory")
	}

	presetPath := s.getPresetPath(moduleName, presetName)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		s.sendPresetSaveErrorEvent(
			moduleName, presetName, "marshal failed", err.Error(),
		)

		return ctxerrors.Wrap(err, "failed to marshal preset data")
	}

	if err := os.WriteFile(presetPath, jsonData, filePerms); err != nil {
		s.sendPresetSaveErrorEvent(
			moduleName, presetName, "write failed", err.Error(),
		)

		return ctxerrors.Wrap(err, "failed to write preset file")
	}

	return nil
}

func (s *PIrateRF) handlePresetRename(
	_ wshub.Hub,
	_ *wshub.Client,
	event *dabluveees.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("Preset rename requested")

	var msg presetRenameMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		s.sendPresetRenameErrorEvent("", "", "", "invalid request", err.Error())

		return nil
	}

	if msg.ModuleName == "" || msg.OldName == "" || msg.NewName == "" {
		s.sendPresetRenameErrorEvent(
			msg.ModuleName, msg.OldName, msg.NewName,
			"invalid request", "all names are required",
		)

		return nil
	}

	if err := s.renamePresetFile(
		msg.ModuleName, msg.OldName, msg.NewName,
	); err != nil {
		return err
	}

	s.sendPresetRenameSuccessEvent(msg.ModuleName, msg.OldName, msg.NewName)

	return nil
}

func (s *PIrateRF) renamePresetFile(
	moduleName, oldName, newName string,
) error {
	oldPath := s.getPresetPath(moduleName, oldName)
	newPath := s.getPresetPath(moduleName, newName)

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		s.sendPresetRenameErrorEvent(
			moduleName, oldName, newName,
			"preset not found", "original preset does not exist",
		)

		return ctxerrors.Wrap(err, "preset file not found")
	}

	if _, err := os.Stat(newPath); err == nil {
		errMsg := "a preset with the new name already exists"
		s.sendPresetRenameErrorEvent(
			moduleName, oldName, newName, "preset exists", errMsg,
		)

		return ctxerrors.New(errMsg)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		s.sendPresetRenameErrorEvent(
			moduleName, oldName, newName, "rename failed", err.Error(),
		)

		return ctxerrors.Wrap(err, "failed to rename preset file")
	}

	return nil
}

func (s *PIrateRF) handlePresetDelete(
	_ wshub.Hub,
	_ *wshub.Client,
	event *dabluveees.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("Preset delete requested")

	var msg presetDeleteMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		s.sendPresetDeleteErrorEvent("", "", "invalid request", err.Error())

		return nil
	}

	if msg.ModuleName == "" || msg.PresetName == "" {
		s.sendPresetDeleteErrorEvent(
			msg.ModuleName, msg.PresetName, "invalid request",
			"module name and preset name are required",
		)

		return nil
	}

	if err := s.deletePresetFile(msg.ModuleName, msg.PresetName); err != nil {
		return err
	}

	s.sendPresetDeleteSuccessEvent(msg.ModuleName, msg.PresetName)

	return nil
}

func (s *PIrateRF) deletePresetFile(moduleName, presetName string) error {
	presetPath := s.getPresetPath(moduleName, presetName)

	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		s.sendPresetDeleteErrorEvent(
			moduleName, presetName,
			"preset not found", "preset does not exist",
		)

		return ctxerrors.Wrap(err, "preset file not found")
	}

	if err := os.Remove(presetPath); err != nil {
		s.sendPresetDeleteErrorEvent(
			moduleName, presetName, "delete failed", err.Error(),
		)

		return ctxerrors.Wrap(err, "failed to delete preset file")
	}

	return nil
}

// Helper functions

func (s *PIrateRF) getPresetPath(moduleName, presetName string) string {
	// Ensure .json extension
	if !strings.HasSuffix(presetName, constants.FileExtensionJSON) {
		presetName += constants.FileExtensionJSON
	}

	return filepath.Join(s.config.FilesDir, presetsDir, moduleName, presetName)
}

// Event sending functions

func (s *PIrateRF) sendPresetLoadSuccessEvent(
	moduleName, presetName string,
	data map[string]any,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetLoadSuccess,
		presetLoadSuccessMessageData{
			ModuleName: moduleName,
			PresetName: presetName,
			Data:       data,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetLoadErrorEvent(
	moduleName, presetName, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetLoadError,
		presetLoadErrorMessageData{
			ModuleName: moduleName,
			PresetName: presetName,
			Error:      errorType,
			Message:    message,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetSaveSuccessEvent(moduleName, presetName string) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetSaveSuccess,
		presetSaveSuccessMessageData{
			ModuleName: moduleName,
			PresetName: presetName,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetSaveErrorEvent(
	moduleName, presetName, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetSaveError,
		presetSaveErrorMessageData{
			ModuleName: moduleName,
			PresetName: presetName,
			Error:      errorType,
			Message:    message,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetRenameSuccessEvent(
	moduleName, oldName, newName string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetRenameSuccess,
		presetRenameSuccessMessageData{
			ModuleName: moduleName,
			OldName:    oldName,
			NewName:    newName,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetRenameErrorEvent(
	moduleName, oldName, newName, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetRenameError,
		presetRenameErrorMessageData{
			ModuleName: moduleName,
			OldName:    oldName,
			NewName:    newName,
			Error:      errorType,
			Message:    message,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetDeleteSuccessEvent(moduleName, presetName string) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetDeleteSuccess,
		presetDeleteSuccessMessageData{
			ModuleName: moduleName,
			PresetName: presetName,
			Timestamp:  time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendPresetDeleteErrorEvent(
	moduleName, presetName, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypePresetDeleteError,
		presetDeleteErrorMessageData{
			ModuleName: moduleName,
			PresetName: presetName,
			Error:      errorType,
			Message:    message,
			Timestamp:  time.Now().Unix(),
		},
	))
}
