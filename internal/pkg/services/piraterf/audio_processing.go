package piraterf

import (
	"context"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/psyb0t/common-go/constants"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

const (
	// Audio conversion timeouts.
	audioConversionTimeout = 30 * time.Second
	audioPlaylistTimeout   = 120 * time.Second
	audioArgsReservedCount = 6 // Reserved space for audio command arguments
)

// audioConversionPostprocessor converts uploaded audio files to optimal format
// using ffmpeg.
func (s *PIrateRF) audioConversionPostprocessor(
	response map[string]any,
) (map[string]any, error) {
	// Get the file path from the response
	filePath, ok := response["path"].(string)
	if !ok {
		return response, nil // Not a string path, return unchanged
	}

	// Check if it's an audio file - convert ALL audio files to ensure proper
	// format
	ext := strings.ToLower(filepath.Ext(filePath))
	audioExtensions := []string{
		constants.FileExtensionMP3,
		constants.FileExtensionM4A,
		constants.FileExtensionFlac,
		constants.FileExtensionOGG,
		constants.FileExtensionAAC,
		constants.FileExtensionWMA,
		constants.FileExtensionWAV,
		".webm",
	}

	isAudioFile := slices.Contains(audioExtensions, ext)

	if !isAudioFile {
		// Not an audio file, return original response unchanged
		return response, nil
	}

	// Convert the audio file
	convertedPath, converted, err := s.convertAudioFileWithFFmpeg(filePath)
	if err != nil {
		logrus.WithError(err).
			WithField("file", filePath).
			Error("Audio conversion failed")

		return response, ctxerrors.Wrapf(err, "audio conversion failed")
	}

	if !converted {
		// No conversion happened, return original response
		return response, nil
	}

	// Clean up the original file
	if removeErr := os.Remove(filePath); removeErr != nil {
		logrus.WithError(removeErr).
			WithField("file", filePath).
			Error("Failed to remove original file")
	}

	// Update response with converted file information
	newResponse := make(map[string]any)
	maps.Copy(newResponse, response)

	// Update path and filename
	newResponse["path"] = convertedPath
	newResponse["saved_filename"] = filepath.Base(convertedPath)
	newResponse["converted"] = true

	// Update file size
	if stat, err := os.Stat(convertedPath); err == nil {
		newResponse["size"] = stat.Size()
	}

	logrus.WithFields(logrus.Fields{
		"original":  filePath,
		"converted": convertedPath,
	}).Info("Audio file converted")

	return newResponse, nil
}

// convertAudioFileWithFFmpeg converts audio file to optimal format using
// ffmpeg. Returns: convertedPath, wasConverted, error.
func (s *PIrateRF) convertAudioFileWithFFmpeg(
	inputPath string,
) (string, bool, error) {
	// Ensure files directory structure exists
	if err := s.ensureFilesDirsExist(); err != nil {
		return "", false, err
	}

	audioUploadsDir := path.Join(s.config.FilesDir, audioUploadsPath)

	// Generate output path in ./files/audio/uploads with .wav extension
	baseFilename := strings.TrimSuffix(
		filepath.Base(inputPath),
		filepath.Ext(inputPath),
	)
	outputPath := filepath.Join(
		audioUploadsDir,
		baseFilename+constants.FileExtensionWAV,
	)

	// Use ffmpeg to convert to optimal format: 16-bit 48kHz mono WAV

	ctx, cancel := context.WithTimeout(
		s.serviceCtx,
		audioConversionTimeout,
	)
	defer cancel()

	// ffmpeg -i input.webm -ar 48000 -ac 1 -c:a pcm_s16le output.wav
	process, err := s.commander.Start(ctx, "ffmpeg", []string{
		"-i", inputPath,
		"-ar", audioSampleRate, // 48kHz sample rate
		"-ac", audioChannels, // mono (1 channel)
		"-c:a", "pcm_s16le", // 16-bit signed little-endian PCM
		"-y", // overwrite output file
		outputPath,
	})
	if err != nil {
		return "", false, ctxerrors.Wrapf(err, "start ffmpeg conversion")
	}

	// Wait for conversion to complete
	if err := process.Wait(); err != nil {
		return "", false, ctxerrors.Wrapf(err, "ffmpeg conversion failed")
	}

	// Verify the output file was created
	if _, err := os.Stat(outputPath); err != nil {
		return "", false, ctxerrors.Wrapf(err, "converted file not found")
	}

	return outputPath, true, nil
}

// createPlaylistFromFiles concatenates multiple audio files into a single
// playlist file using sox. Returns the path to the created playlist file.
// If outputDir is empty, uses the default uploads directory for permanent
// playlists. If outputDir is specified, uses that directory for temporary
// playlists.
func (s *PIrateRF) createPlaylistFromFiles(
	playlistName string,
	filePaths []string,
	outputDir ...string,
) (string, error) {
	outputPath := s.getPlaylistOutputPath(playlistName, outputDir...)
	if outputPath == "" {
		return "", ctxerrors.New("failed to determine output path")
	}

	return s.executePlaylistCreation(filePaths, outputPath)
}

func (s *PIrateRF) getPlaylistOutputPath(
	playlistName string,
	outputDir ...string,
) string {
	playlistName = s.ensureWavExtension(playlistName)

	if len(outputDir) > 0 && outputDir[0] != "" {
		return filepath.Join(outputDir[0], playlistName)
	}

	if err := s.ensureFilesDirsExist(); err != nil {
		return ""
	}

	audioUploadsDir := path.Join(s.config.FilesDir, audioUploadsPath)

	return filepath.Join(audioUploadsDir, playlistName)
}

func (s *PIrateRF) ensureWavExtension(playlistName string) string {
	if !strings.HasSuffix(playlistName, constants.FileExtensionWAV) {
		return playlistName + constants.FileExtensionWAV
	}

	return playlistName
}

func (s *PIrateRF) executePlaylistCreation(
	filePaths []string,
	outputPath string,
) (string, error) {
	ctx, cancel := context.WithTimeout(s.serviceCtx, audioPlaylistTimeout)
	defer cancel()

	soxArgs := s.buildSoxArgs(filePaths, outputPath)

	process, err := s.commander.Start(ctx, "sox", soxArgs)
	if err != nil {
		return "", ctxerrors.Wrapf(err, "start sox playlist creation")
	}

	if err := process.Wait(); err != nil {
		return "", ctxerrors.Wrapf(err, "sox playlist creation failed")
	}

	if _, err := os.Stat(outputPath); err != nil {
		return "", ctxerrors.Wrapf(err, "playlist file not found")
	}

	return outputPath, nil
}

func (s *PIrateRF) buildSoxArgs(
	filePaths []string,
	outputPath string,
) []string {
	soxArgs := make([]string, 0, len(filePaths)+audioArgsReservedCount)
	soxArgs = append(soxArgs, filePaths...)
	soxArgs = append(soxArgs, "-r", audioSampleRate)
	soxArgs = append(soxArgs, "-b", audioBitDepth)
	soxArgs = append(soxArgs, "-c", audioChannels)
	soxArgs = append(soxArgs, outputPath)

	return soxArgs
}
