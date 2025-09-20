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

	msg, err := s.parseExecutionMessage(event, logger)
	if err != nil {
		return err
	}

	if err := s.validateModuleInDev(msg.ModuleName, logger); err != nil {
		return s.handleModuleValidationError(err, msg.ModuleName, logger)
	}

	return s.processModuleExecution(msg, client, logger)
}

func (s *PIrateRF) parseExecutionMessage(
	event *websocket.Event,
	logger *logrus.Entry,
) (*rpitxExecutionStartMessage, error) {
	var msg rpitxExecutionStartMessage
	if err := json.Unmarshal(event.Data, &msg); err != nil {
		logger.WithError(err).Error("failed to unmarshal RPITX exec message")

		return nil, ctxerrors.Wrap(err, "failed to unmarshal RPITX exec message")
	}

	return &msg, nil
}

func (s *PIrateRF) handleModuleValidationError(err error, moduleName gorpitx.ModuleName, logger *logrus.Entry) error {
	logger.WithError(err).
		WithField("module", moduleName).
		Error("Module validation failed - unsupported module")

	s.executionManager.SendError(
		"unknown module",
		fmt.Sprintf("%s: unknown module", moduleName),
	)

	return ctxerrors.Wrap(err, "module validation failed")
}

func (s *PIrateRF) processModuleExecution(
	msg *rpitxExecutionStartMessage,
	client *websocket.Client,
	logger *logrus.Entry,
) error {
	finalTimeout := msg.Timeout
	finalArgs := msg.Args

	switch msg.ModuleName {
	case gorpitx.ModuleNamePIFMRDS:
		return s.handlePIFMRDSExecution(msg, finalTimeout, client, logger)
	case gorpitx.ModuleNameSPECTRUMPAINT:
		return s.handleSPECTRUMPAINTExecution(msg, finalTimeout, client, logger)
	case gorpitx.ModuleNamePICHIRP:
		return s.handlePICHIRPExecution(msg, finalTimeout, client, logger)
	default:
		return s.executionManager.startExecution(s.serviceCtx, msg.ModuleName, finalArgs, finalTimeout, client, nil)
	}
}

func (s *PIrateRF) handlePIFMRDSExecution(
	msg *rpitxExecutionStartMessage,
	finalTimeout int,
	client *websocket.Client,
	logger *logrus.Entry,
) error {
	processedTimeout, cleanupPath, finalArgs, err := s.processAudioModifications(*msg, finalTimeout, logger)
	if err != nil {
		logger.WithError(err).Error("Audio processing failed")

		return ctxerrors.Wrap(err, "audio processing failed")
	}

	callback := s.createCleanupCallback(cleanupPath, logger)

	return s.executionManager.startExecution(s.serviceCtx, msg.ModuleName, finalArgs, processedTimeout, client, callback)
}

func (s *PIrateRF) handleSPECTRUMPAINTExecution(
	msg *rpitxExecutionStartMessage,
	finalTimeout int,
	client *websocket.Client,
	logger *logrus.Entry,
) error {
	modifiedArgs, err := s.processImageModifications(msg.Args, logger)
	if err != nil {
		logger.WithError(err).Error("Image processing failed")

		return ctxerrors.Wrap(err, "image processing failed")
	}

	return s.executionManager.startExecution(s.serviceCtx, msg.ModuleName, modifiedArgs, finalTimeout, client, nil)
}

func (s *PIrateRF) handlePICHIRPExecution(
	msg *rpitxExecutionStartMessage,
	finalTimeout int,
	client *websocket.Client,
	logger *logrus.Entry,
) error {
	logger.Debug("Processing PICHIRP execution request")

	return s.executionManager.startExecution(s.serviceCtx, msg.ModuleName, msg.Args, finalTimeout, client, nil)
}

func (s *PIrateRF) createCleanupCallback(cleanupPath string, logger *logrus.Entry) func() error {
	if cleanupPath == "" {
		return nil
	}

	return func() error {
		if err := os.Remove(cleanupPath); err != nil {
			logger.WithError(err).WithField("path", cleanupPath).Warn("Failed to cleanup temporary audio file")

			return ctxerrors.Wrap(err, "failed to remove temporary audio file")
		}

		logger.WithField("path", cleanupPath).Debug("Cleaned up temporary audio file")

		return nil
	}
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
	stdout, stderr, err := s.commander.Output(
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
	if !msg.PlayOnce {
		return audioFile, existingTempPath, modifiedArgs, nil
	}

	silenceAudioPath := s.generateSilenceFilePath()
	if err := s.addSilenceToAudio(audioFile, silenceAudioPath, logger); err != nil {
		return audioFile, existingTempPath, modifiedArgs, err
	}

	finalArgs, err := s.updateArgsWithSilenceFile(modifiedArgs, silenceAudioPath)
	if err != nil {
		return audioFile, existingTempPath, modifiedArgs, err
	}

	return silenceAudioPath, silenceAudioPath, finalArgs, nil
}

func (s *PIrateRF) generateSilenceFilePath() string {
	playlistID := uuid.New().String()

	return "/tmp/" + playlistID + "_with_silence" + constants.FileExtensionWAV
}

func (s *PIrateRF) addSilenceToAudio(audioFile, silenceAudioPath string, logger *logrus.Entry) error {
	logger.WithFields(logrus.Fields{
		"originalAudio": audioFile,
		"silenceFile":   silenceAudioPath,
	}).Debug("Creating audio file with 2 seconds of silence for Play Once mode")

	ctx, cancel := context.WithTimeout(s.serviceCtx, audioConversionTimeout)
	defer cancel()

	process, err := s.commander.Start(ctx, "sox", []string{
		audioFile,
		silenceAudioPath,
		"pad", "0", "2",
	})
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to start sox silence addition")
	}

	if err := process.Wait(); err != nil {
		return ctxerrors.Wrapf(err, "sox silence addition failed")
	}

	if _, err := os.Stat(silenceAudioPath); err != nil {
		return ctxerrors.Wrapf(err, "silence audio file not found")
	}

	logger.WithField("silenceAudioPath", silenceAudioPath).
		Debug("Successfully created audio file with silence")

	return nil
}

func (s *PIrateRF) updateArgsWithSilenceFile(
	modifiedArgs json.RawMessage,
	silenceAudioPath string,
) (json.RawMessage, error) {
	var argsMap map[string]any
	if err := json.Unmarshal(modifiedArgs, &argsMap); err != nil {
		return modifiedArgs, ctxerrors.Wrapf(err, "failed to unmarshal args")
	}

	argsMap["audio"] = silenceAudioPath

	finalArgs, err := json.Marshal(argsMap)
	if err != nil {
		return modifiedArgs, ctxerrors.Wrapf(err, "failed to marshal final args")
	}

	return finalArgs, nil
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
