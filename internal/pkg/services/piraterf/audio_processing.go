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

	"github.com/psyb0t/commander"
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

// fileConversionPostprocessor handles both audio and image file conversions
// based on file type.
func (s *PIrateRF) fileConversionPostprocessor(
	response map[string]any,
) (map[string]any, error) {
	// Try audio conversion first
	audioResponse, err := s.audioConversionPostprocessor(response)
	if err != nil {
		return response, err
	}

	// If audio conversion happened, return that result
	if converted, ok := audioResponse["converted"].(bool); ok && converted {
		return audioResponse, nil
	}

	// Try image conversion
	return s.imageConversionPostprocessor(response)
}

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
	}).Info("Audio file converted successfully")

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
	cmd := commander.New()

	ctx, cancel := context.WithTimeout(
		s.serviceCtx,
		audioConversionTimeout,
	)
	defer cancel()

	// ffmpeg -i input.webm -ar 48000 -ac 1 -c:a pcm_s16le output.wav
	process, err := cmd.Start(ctx, "ffmpeg", []string{
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
	var outputPath string

	if len(outputDir) > 0 && outputDir[0] != "" {
		// Use specified directory (for temporary playlists)
		// Ensure the playlist name ends with .wav
		if !strings.HasSuffix(
			playlistName,
			constants.FileExtensionWAV,
		) {
			playlistName += constants.FileExtensionWAV
		}

		outputPath = filepath.Join(outputDir[0], playlistName)
	} else {
		// Use default uploads directory (for permanent playlists)
		// Ensure files directory structure exists
		if err := s.ensureFilesDirsExist(); err != nil {
			return "", err
		}

		// Generate output path in uploads directory
		audioUploadsDir := path.Join(
			s.config.FilesDir,
			audioUploadsPath,
		)

		// Ensure the playlist name ends with .wav
		if !strings.HasSuffix(
			playlistName,
			constants.FileExtensionWAV,
		) {
			playlistName += constants.FileExtensionWAV
		}

		outputPath = filepath.Join(audioUploadsDir, playlistName)
	}

	// Use sox to concatenate files with consistent format
	cmd := commander.New()

	ctx, cancel := context.WithTimeout(
		s.serviceCtx,
		audioPlaylistTimeout, // Longer timeout for playlist creation
	)
	defer cancel()

	// sox input1.wav input2.wav input3.wav -r 48000 -b 16 -c 1 output.wav
	soxArgs := make(
		[]string,
		0,
		len(filePaths)+audioArgsReservedCount,
	)
	soxArgs = append(soxArgs, filePaths...)
	soxArgs = append(soxArgs, "-r", audioSampleRate)
	soxArgs = append(soxArgs, "-b", audioBitDepth)
	soxArgs = append(soxArgs, "-c", audioChannels)
	soxArgs = append(soxArgs, outputPath)

	process, err := cmd.Start(ctx, "sox", soxArgs)
	if err != nil {
		return "", ctxerrors.Wrapf(err, "start sox playlist creation")
	}

	// Wait for concatenation to complete
	if err := process.Wait(); err != nil {
		return "", ctxerrors.Wrapf(err, "sox playlist creation failed")
	}

	// Verify the output file was created
	if _, err := os.Stat(outputPath); err != nil {
		return "", ctxerrors.Wrapf(err, "playlist file not found")
	}

	return outputPath, nil
}
