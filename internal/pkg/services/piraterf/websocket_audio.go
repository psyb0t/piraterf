package piraterf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/common-go/constants"
	"github.com/sirupsen/logrus"
)

const (
	eventTypeAudioPlaylistCreate = dabluveees.EventType(
		"audio.playlist.create",
	)
	eventTypeAudioPlaylistCreateSuccess = dabluveees.EventType(
		"audio.playlist.create.success",
	)
	eventTypeAudioPlaylistCreateError = dabluveees.EventType(
		"audio.playlist.create.error",
	)
)

type audioPlaylistCreateMessage struct {
	PlaylistFileName string   `json:"playlistFileName"` // Name for the output file
	Files            []string `json:"files"`            // Array of full file paths
}

type audioPlaylistCreateSuccessMessageData struct {
	PlaylistName string `json:"playlistName"`
	FilePath     string `json:"filePath"`
	Timestamp    int64  `json:"timestamp"`
}

type audioPlaylistCreateErrorMessageData struct {
	PlaylistName string `json:"playlistName"`
	Error        string `json:"error"`
	Message      string `json:"message"`
	Timestamp    int64  `json:"timestamp"`
}

func (s *PIrateRF) handleAudioPlaylistCreate(
	_ wshub.Hub,
	_ *wshub.Client,
	event *dabluveees.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("Playlist creation requested")

	var msg audioPlaylistCreateMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		logger.WithError(err).Error(
			"failed to unmarshal audio playlist create message",
		)
		s.sendAudioPlaylistCreateErrorEvent(
			msg.PlaylistFileName,
			"invalid request",
			err.Error(),
		)

		return nil
	}

	filePaths, valid := s.validatePlaylistRequest(msg)
	if !valid {
		return nil
	}

	// Create playlist by concatenating all files
	outputPath, err := s.createPlaylistFromFiles(msg.PlaylistFileName, filePaths)
	if err != nil {
		logger.WithError(err).Error("failed to create playlist")
		s.sendAudioPlaylistCreateErrorEvent(
			msg.PlaylistFileName,
			"creation failed",
			err.Error(),
		)

		return nil
	}

	logger.Infof("Audio playlist created successfully: %s", outputPath)
	s.sendAudioPlaylistCreateSuccessEvent(msg.PlaylistFileName, outputPath)

	return nil
}

// validatePlaylistRequest validates the playlist creation request and returns
// converted file paths if validation passes.
func (s *PIrateRF) validatePlaylistRequest(
	msg audioPlaylistCreateMessage,
) ([]string, bool) {
	if len(msg.Files) == 0 {
		s.sendAudioPlaylistCreateErrorEvent(
			msg.PlaylistFileName,
			"empty playlist",
			"no files provided",
		)

		return nil, false
	}

	// Convert HTTP paths to file system paths first
	filePaths := make([]string, len(msg.Files))
	for i, httpPath := range msg.Files {
		filePaths[i] = s.convertHTTPPathToFileSystem(httpPath)
	}

	// Validate all files exist using converted paths
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			s.sendAudioPlaylistCreateErrorEvent(
				msg.PlaylistFileName,
				"file not found",
				fmt.Sprintf("file does not exist: %s", filePath),
			)

			return nil, false
		}
	}

	return filePaths, true
}

// Event sending functions for audio playlist operations.
func (s *PIrateRF) sendAudioPlaylistCreateSuccessEvent(
	playlistName, filePath string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypeAudioPlaylistCreateSuccess,
		audioPlaylistCreateSuccessMessageData{
			PlaylistName: playlistName,
			FilePath:     filePath,
			Timestamp:    time.Now().Unix(),
		},
	))
}

func (s *PIrateRF) sendAudioPlaylistCreateErrorEvent(
	playlistName, errorType, message string,
) {
	s.websocketHub.BroadcastToAll(dabluveees.NewEvent(
		eventTypeAudioPlaylistCreateError,
		audioPlaylistCreateErrorMessageData{
			PlaylistName: playlistName,
			Error:        errorType,
			Message:      message,
			Timestamp:    time.Now().Unix(),
		},
	))
}

// convertHTTPPathToFileSystem converts HTTP paths like
// "/files/audio/sfx/file.wav"
// to file system paths like "./files/audio/sfx/file.wav".
func (s *PIrateRF) convertHTTPPathToFileSystem(httpPath string) string {
	// Remove leading "/files" and prepend with the actual files directory
	if after, ok := strings.CutPrefix(httpPath, "/files/"); ok {
		relativePath := after

		return filepath.Join(s.config.FilesDir, relativePath)
	}

	// If it does not start with /files/, return as-is (might already be a
	// filesystem path)
	return httpPath
}
