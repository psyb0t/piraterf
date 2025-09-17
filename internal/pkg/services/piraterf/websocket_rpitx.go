package piraterf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/constants"
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
)

const (
	eventTypeRPITXExecutionStart      websocket.EventType = "rpitx.execution.start"
	eventTypeRPITXExecutionStarted    websocket.EventType = "rpitx.execution.started"
	eventTypeRPITXExecutionStop       websocket.EventType = "rpitx.execution.stop"
	eventTypeRPITXExecutionStopped    websocket.EventType = "rpitx.execution.stopped"
	eventTypeRPITXExecutionError      websocket.EventType = "rpitx.execution.error"
	eventTypeRPITXExecutionOutputLine websocket.EventType = "rpitx.execution.output-line"

	// Audio duration rounding offset for converting float to int
	// seconds.
	durationRoundingOffset = 0.5
)

type rpitxExecutionStartMessage struct {
	ModuleName gorpitx.ModuleName `json:"moduleName"`
	Args       json.RawMessage    `json:"args"`
	Timeout    int                `json:"timeout"`  // timeout in seconds
	PlayOnce   bool               `json:"playOnce"` // use duration as timeout
	Intro      *string            `json:"intro"`    // intro file path (optional)
	Outro      *string            `json:"outro"`    // outro file path (optional)
}

type rpitxExecutionStartedMessageData struct {
	ModuleName         gorpitx.ModuleName `json:"moduleName"`
	Args               json.RawMessage    `json:"args"`
	InitiatingClientID string             `json:"initiatingClientId"`
	Timestamp          int64              `json:"timestamp"`
}

type rpitxExecutionStoppedMessageData struct {
	InitiatingClientID string `json:"initiatingClientId"`
	StoppingClientID   string `json:"stoppingClientId"`
	Timestamp          int64  `json:"timestamp"`
}

type rpitxExecutionErrorMessageData struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type rpitxExecutionOutputLineMessageData struct {
	Type      string `json:"type"`
	Line      string `json:"line"`
	Timestamp int64  `json:"timestamp"`
}

//nolint:funlen
func (s *PIrateRF) handleRPITXExecutionStart(
	_ websocket.Hub,
	client *websocket.Client,
	event *websocket.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("RPITX execution start requested")

	var msg rpitxExecutionStartMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		logger.WithError(err).Error("failed to unmarshal RPITX exec message")

		return ctxerrors.Wrap(err, "failed to unmarshal RPITX exec message")
	}

	// Early module validation in dev mode
	if err := s.validateModuleInDev(msg.ModuleName, logger); err != nil {
		logger.WithError(err).
			WithField("module", msg.ModuleName).
			Error("Module validation failed - unsupported module")

		// Send error event to UI as well
		s.executionManager.SendError(
			"unknown module",
			fmt.Sprintf("%s: unknown module", msg.ModuleName),
		)

		return ctxerrors.Wrap(err, "module validation failed")
	}

	finalTimeout := msg.Timeout

	finalArgs := msg.Args

	// Handle audio modifications for pifmrds module
	if msg.ModuleName == gorpitx.ModuleNamePIFMRDS {
		var (
			err         error
			cleanupPath string
		)

		finalTimeout, cleanupPath, finalArgs, err = s.processAudioModifications(
			msg,
			finalTimeout,
			logger,
		)
		if err != nil {
			logger.WithError(err).Error("Audio processing failed")

			return ctxerrors.Wrap(err, "audio processing failed")
		}

		// Create callback for temporary file cleanup
		var callback func() error
		if cleanupPath != "" {
			callback = func() error {
				if err := os.Remove(cleanupPath); err != nil {
					logger.WithError(err).
						WithField("path", cleanupPath).
						Warn("Failed to cleanup temporary audio file")

					return ctxerrors.Wrap(err, "failed to remove temporary audio file")
				}

				logger.WithField("path", cleanupPath).
					Debug("Cleaned up temporary audio file")

				return nil
			}
		}

		return s.executionManager.startExecution(
			s.serviceCtx,
			msg.ModuleName,
			finalArgs,
			finalTimeout,
			client,
			callback,
		)
	}

	// Handle image modifications for spectrumpaint module
	if msg.ModuleName == gorpitx.ModuleNameSPECTRUMPAINT {
		modifiedArgs, err := s.processImageModifications(
			msg.Args,
			logger,
		)
		if err != nil {
			logger.WithError(err).Error("Image processing failed")

			return ctxerrors.Wrap(err, "image processing failed")
		}

		finalArgs = modifiedArgs
	}

	return s.executionManager.startExecution(
		s.serviceCtx,
		msg.ModuleName,
		finalArgs,
		finalTimeout,
		client,
		nil,
	)
}

func (s *PIrateRF) handleRPITXExecutionStop(
	_ websocket.Hub,
	client *websocket.Client,
	event *websocket.Event,
) error {
	logger := logrus.WithFields(logrus.Fields{
		constants.FieldEventType: event.Type,
		constants.FieldEventID:   event.ID,
	})

	logger.Debug("RPITX execution stop requested")

	return s.executionManager.stopExecution(client)
}

func (s *PIrateRF) getAudioDurationWithSox(audioFile string) (float64, error) {
	// Use sox to get audio duration
	cmd := commander.New()

	stdout, stderr, err := cmd.Output(
		s.serviceCtx,
		"sox",
		[]string{"--info", "-D", audioFile},
	)
	if err != nil {
		return 0, ctxerrors.Wrapf(
			err,
			"sox command failed, stderr: %s",
			string(stderr),
		)
	}

	// Parse duration from sox output
	durationStr := strings.TrimSpace(string(stdout))

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, ctxerrors.Wrapf(err, "failed to parse duration '%s'", durationStr)
	}

	return duration, nil
}

func (s *PIrateRF) processAudioModifications(
	msg rpitxExecutionStartMessage,
	originalTimeout int,
	logger *logrus.Entry,
) (int, string, json.RawMessage, error) {
	var argsMap map[string]any
	if err := json.Unmarshal(msg.Args, &argsMap); err != nil {
		return originalTimeout, "", msg.Args,
			ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	audioFile, ok := argsMap["audio"].(string)
	if !ok || audioFile == "" {
		return originalTimeout, "", msg.Args, nil
	}

	// Handle intro/outro playlist creation
	tempPlaylistPath, modifiedArgs, err := s.processIntroOutro(
		msg,
		argsMap,
		audioFile,
		logger,
	)
	if err != nil {
		return originalTimeout, "", msg.Args, err
	}

	// Update audioFile to use playlist if one was created
	if tempPlaylistPath != "" {
		audioFile = tempPlaylistPath
	}

	// Add silence for Play Once mode
	finalAudioFile, cleanupPaths, finalArgs, err := s.processPlayOnceSilence(
		msg,
		audioFile,
		tempPlaylistPath,
		modifiedArgs,
		logger,
	)
	if err != nil {
		return originalTimeout, tempPlaylistPath, modifiedArgs, err
	}

	// Handle Play Once timeout calculation
	finalTimeout, err := s.processPlayOnceTimeout(
		msg,
		finalAudioFile,
		originalTimeout,
		logger,
	)
	if err != nil {
		return originalTimeout, cleanupPaths, finalArgs, err
	}

	return finalTimeout, cleanupPaths, finalArgs, nil
}

func (s *PIrateRF) processIntroOutro(
	msg rpitxExecutionStartMessage,
	argsMap map[string]any,
	audioFile string,
	logger *logrus.Entry,
) (string, json.RawMessage, error) {
	// Create playlist if intro/outro specified
	if msg.Intro == nil && msg.Outro == nil {
		return "", msg.Args, nil
	}

	playlistPath, err := s.createTempPlaylist(
		audioFile,
		msg.Intro,
		msg.Outro,
		logger,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to create temp playlist")

		return "", msg.Args, ctxerrors.Wrap(err, "failed to create temp playlist")
	}

	// Update the audio file path in args to use the playlist
	argsMap["audio"] = playlistPath

	modifiedArgs, err := json.Marshal(argsMap)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal modified args")

		return "", msg.Args, ctxerrors.Wrap(err, "failed to marshal modified args")
	}

	logger.WithFields(logrus.Fields{
		"originalAudio": audioFile,
		"playlistPath":  playlistPath,
	}).Debug("Created temporary playlist")

	return playlistPath, modifiedArgs, nil
}

func (s *PIrateRF) processPlayOnceTimeout(
	msg rpitxExecutionStartMessage,
	audioFile string,
	originalTimeout int,
	logger *logrus.Entry,
) (int, error) {
	// Handle Play Once timeout calculation
	if !msg.PlayOnce {
		return originalTimeout, nil
	}

	duration, err := s.getAudioDurationWithSox(audioFile)
	if err != nil {
		logger.WithError(err).Error("Play Once: failed to get audio duration")

		return originalTimeout, ctxerrors.Wrap(err, "failed to get audio duration")
	}

	audioDurationSeconds := int(duration + durationRoundingOffset) // Round up
	logger.WithFields(logrus.Fields{
		"audioFile":       audioFile,
		"duration":        duration,
		"originalTimeout": msg.Timeout,
	}).Debug("Play Once: got audio duration")

	// If no timeout set (0), use audio duration
	if msg.Timeout == 0 {
		logger.WithField("finalTimeout", audioDurationSeconds).
			Debug("Play Once: using audio duration (no timeout set)")

		return audioDurationSeconds, nil
	}

	// Use the lower of audio duration or user timeout
	if audioDurationSeconds < msg.Timeout {
		logger.WithField("finalTimeout", audioDurationSeconds).Debug(
			"Play Once: using audio duration (shorter than timeout)")

		return audioDurationSeconds, nil
	}

	logger.WithField("finalTimeout", msg.Timeout).
		Debug("Play Once: using user timeout (shorter than audio)")

	return msg.Timeout, nil
}

func (s *PIrateRF) createTempPlaylist(
	mainAudio string,
	intro, outro *string,
	logger *logrus.Entry,
) (string, error) {
	// Generate unique filename for temporary playlist
	playlistID := uuid.New().String()
	playlistName := playlistID + constants.FileExtensionWAV

	// Build file paths array for concatenation
	var filePaths []string

	// Add intro if specified
	if intro != nil && *intro != "" {
		filePaths = append(filePaths, *intro)
	}

	// Add main audio file
	filePaths = append(filePaths, mainAudio)

	// Add outro if specified
	if outro != nil && *outro != "" {
		filePaths = append(filePaths, *outro)
	}

	logger.WithFields(logrus.Fields{
		"intro":     intro,
		"mainAudio": mainAudio,
		"outro":     outro,
		"filePaths": filePaths,
	}).Debug("Creating temporary playlist using existing createPlaylistFromFiles")

	// Use existing function with /tmp directory for temporary playlist
	playlistPath, err := s.createPlaylistFromFiles(playlistName, filePaths, "/tmp")
	if err != nil {
		return "", ctxerrors.Wrapf(err, "failed to create temporary playlist")
	}

	logger.WithField("playlistPath", playlistPath).
		Debug("Successfully created temporary playlist")

	return playlistPath, nil
}

// processPlayOnceSilence adds 2 seconds of silence to audio file for Play
// Once mode.
func (s *PIrateRF) processPlayOnceSilence(
	msg rpitxExecutionStartMessage,
	audioFile string,
	existingTempPath string,
	modifiedArgs json.RawMessage,
	logger *logrus.Entry,
) (string, string, json.RawMessage, error) {
	// Only add silence if Play Once is enabled
	if !msg.PlayOnce {
		return audioFile, existingTempPath, modifiedArgs, nil
	}

	// Generate unique filename for audio file with silence
	playlistID := uuid.New().String()
	silenceAudioPath := "/tmp/" + playlistID + "_with_silence" +
		constants.FileExtensionWAV

	logger.WithFields(logrus.Fields{
		"originalAudio": audioFile,
		"silenceFile":   silenceAudioPath,
	}).Debug("Creating audio file with 2 seconds of silence for Play Once mode")

	// Use sox to add 2 seconds of silence at the end
	// sox input.wav silence_output.wav pad 0 2
	cmd := commander.New()

	ctx, cancel := context.WithTimeout(s.serviceCtx, audioConversionTimeout)
	defer cancel()

	process, err := cmd.Start(ctx, "sox", []string{
		audioFile,
		silenceAudioPath,
		"pad", "0", "2", // Add 2 seconds of silence at the end
	})
	if err != nil {
		return audioFile, existingTempPath, modifiedArgs,
			ctxerrors.Wrapf(err, "failed to start sox silence addition")
	}

	if err := process.Wait(); err != nil {
		return audioFile,
			existingTempPath,
			modifiedArgs,
			ctxerrors.Wrapf(err, "sox silence addition failed")
	}

	// Verify the output file was created
	if _, err := os.Stat(silenceAudioPath); err != nil {
		return audioFile,
			existingTempPath,
			modifiedArgs,
			ctxerrors.Wrapf(err, "silence audio file not found")
	}

	logger.WithField("silenceAudioPath", silenceAudioPath).
		Debug("Successfully created audio file with silence")

	// Update the args to use the new audio file with silence
	var argsMap map[string]any
	if err := json.Unmarshal(modifiedArgs, &argsMap); err != nil {
		return audioFile,
			existingTempPath,
			modifiedArgs,
			ctxerrors.Wrapf(err, "failed to unmarshal args")
	}

	argsMap["audio"] = silenceAudioPath

	finalArgs, err := json.Marshal(argsMap)
	if err != nil {
		return audioFile,
			existingTempPath,
			modifiedArgs,
			ctxerrors.Wrapf(err, "failed to marshal final args")
	}

	// Return the new file path, cleanup path, and updated args
	return silenceAudioPath, silenceAudioPath, finalArgs, nil
}

// validateModuleInDev validates that a module is supported in development mode.
func (s *PIrateRF) validateModuleInDev(
	moduleName gorpitx.ModuleName,
	logger *logrus.Entry,
) error {
	// Use rpitx instance to check if module is supported
	if s.rpitx.IsSupportedModule(moduleName) {
		logger.WithField("module", moduleName).
			Debug("Module validation passed")

		return nil
	}

	// Get list of supported modules for error reporting
	supportedModules := s.rpitx.GetSupportedModules()

	logger.WithField("module", moduleName).
		WithField("supportedModules", supportedModules).
		Error("Unsupported module requested")

	return ctxerrors.Wrap(gorpitx.ErrUnknownModule, moduleName)
}

// processImageModifications handles image conversion for SPECTRUMPAINT module.
func (s *PIrateRF) processImageModifications(
	args json.RawMessage,
	logger *logrus.Entry,
) (json.RawMessage, error) {
	var argsMap map[string]any
	if err := json.Unmarshal(args, &argsMap); err != nil {
		return args, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	pictureFile, ok := argsMap["pictureFile"].(string)
	if !ok || pictureFile == "" {
		return args, nil
	}

	// Convert image to YUV format if needed
	convertedPath, err := s.convertImageToYUV(pictureFile, logger)
	if err != nil {
		return args, ctxerrors.Wrap(err, "failed to convert image")
	}

	// Update the picture file path in args to use the converted file
	argsMap["pictureFile"] = convertedPath

	modifiedArgs, err := json.Marshal(argsMap)
	if err != nil {
		return args, ctxerrors.Wrap(err, "failed to marshal modified args")
	}

	logger.WithFields(logrus.Fields{
		"originalImage": pictureFile,
		"convertedPath": convertedPath,
	}).Debug("Image converted for SPECTRUMPAINT")

	return modifiedArgs, nil
}
