package piraterf

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/common-go/constants"
	"github.com/sirupsen/logrus"
)

const (
	eventTypeFileRename        websocket.EventType = "file.rename"
	eventTypeFileRenameSuccess websocket.EventType = "file.rename.success"
	eventTypeFileRenameError   websocket.EventType = "file.rename.error"
	eventTypeFileDelete        websocket.EventType = "file.delete"
	eventTypeFileDeleteSuccess websocket.EventType = "file.delete.success"
	eventTypeFileDeleteError   websocket.EventType = "file.delete.error"
)

type fileRenameMessage struct {
	FilePath string `json:"filePath"` // Full path to original file
	NewName  string `json:"newName"`  // Just the new filename
}

type fileRenameSuccessMessageData struct {
	FileName  string `json:"fileName"`
	NewName   string `json:"newName"`
	Timestamp int64  `json:"timestamp"`
}

type fileRenameErrorMessageData struct {
	FileName  string `json:"fileName"`
	NewName   string `json:"newName"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type fileDeleteMessage struct {
	FilePath string `json:"filePath"` // Full path to file
}

type fileDeleteSuccessMessageData struct {
	FileName  string `json:"fileName"`
	Timestamp int64  `json:"timestamp"`
}

type fileDeleteErrorMessageData struct {
	FileName  string `json:"fileName"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func (s *PIrateRF) handleFileRename(
	_ websocket.Hub,
	_ *websocket.Client,
	event *websocket.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("File rename requested")

	var msg fileRenameMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		logger.WithError(err).
			Error("failed to unmarshal file rename message")
		s.sendFileRenameErrorEvent(
			msg.FilePath,
			msg.NewName,
			"invalid request",
			err.Error(),
		)

		return nil
	}

	oldPath, newPath, valid := s.validateFileRenameRequest(msg)
	if !valid {
		return nil
	}

	// Rename the file
	if err := os.Rename(oldPath, newPath); err != nil {
		logger.WithError(err).
			Error("failed to rename file")
		s.sendFileRenameErrorEvent(
			msg.FilePath,
			msg.NewName,
			"rename failed",
			err.Error(),
		)

		return nil
	}

	logger.Infof("File renamed from %s to %s", msg.FilePath, msg.NewName)
	s.sendFileRenameSuccessEvent(msg.FilePath, msg.NewName)

	return nil
}

// validateFileRenameRequest validates the file rename request and returns
// old and new paths if validation passes.
func (s *PIrateRF) validateFileRenameRequest(
	msg fileRenameMessage,
) (string, string, bool) {
	// Use the original file's full path and build new path with same directory
	oldPath := msg.FilePath
	fileDir := path.Dir(oldPath)
	newPath := path.Join(fileDir, msg.NewName)

	// Check if old file exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		logrus.WithFields(logrus.Fields{
			"oldPath": oldPath,
			"newName": msg.NewName,
		}).Error("File rename failed: original file does not exist")

		s.sendFileRenameErrorEvent(
			msg.FilePath,
			msg.NewName,
			"file not found",
			"original file does not exist",
		)

		return "", "", false
	}

	// Check if new file already exists
	if _, err := os.Stat(newPath); err == nil {
		logrus.WithFields(logrus.Fields{
			"oldPath": oldPath,
			"newPath": newPath,
			"newName": msg.NewName,
		}).Error("File rename failed: target file already exists")

		s.sendFileRenameErrorEvent(
			msg.FilePath,
			msg.NewName,
			"file exists",
			"target file already exists",
		)

		return "", "", false
	}

	return oldPath, newPath, true
}

func (s *PIrateRF) handleFileDelete(
	_ websocket.Hub,
	_ *websocket.Client,
	event *websocket.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("File delete requested")

	var msg fileDeleteMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		logger.WithError(err).
			Error("failed to unmarshal file delete message")
		s.sendFileDeleteErrorEvent(msg.FilePath, "invalid request", err.Error())

		return nil
	}

	// Use the full file path directly
	filePath := msg.FilePath

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		s.sendFileDeleteErrorEvent(
			msg.FilePath,
			"file not found",
			"file does not exist",
		)

		return nil
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		logger.WithError(err).
			Error("failed to delete file")
		s.sendFileDeleteErrorEvent(msg.FilePath, "delete failed", err.Error())

		return nil
	}

	logger.Infof("File deleted: %s", msg.FilePath)
	s.sendFileDeleteSuccessEvent(msg.FilePath)

	return nil
}

// Event sending functions for file operations.
func (s *PIrateRF) sendFileRenameSuccessEvent(filePath, newName string) {
	s.websocketHub.BroadcastToAll(websocket.NewEvent(
		eventTypeFileRenameSuccess,
		fileRenameSuccessMessageData{
			FileName:  filePath,
			NewName:   newName,
			Timestamp: time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendFileRenameErrorEvent(
	filePath, newName, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(websocket.NewEvent(
		eventTypeFileRenameError,
		fileRenameErrorMessageData{
			FileName:  filePath,
			NewName:   newName,
			Error:     errorType,
			Message:   message,
			Timestamp: time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendFileDeleteSuccessEvent(filePath string) {
	s.websocketHub.BroadcastToAll(websocket.NewEvent(
		eventTypeFileDeleteSuccess,
		fileDeleteSuccessMessageData{
			FileName:  filePath,
			Timestamp: time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendFileDeleteErrorEvent(
	filePath, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(websocket.NewEvent(
		eventTypeFileDeleteError,
		fileDeleteErrorMessageData{
			FileName:  filePath,
			Error:     errorType,
			Message:   message,
			Timestamp: time.Now().Unix(),
		},
	))
}
