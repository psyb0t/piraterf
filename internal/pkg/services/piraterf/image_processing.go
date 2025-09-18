package piraterf

import (
	"context"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

// imageConversionPostprocessor converts uploaded image files to YUV format
// for SPECTRUMPAINT using ImageMagick.
func (s *PIrateRF) imageConversionPostprocessor(
	response map[string]any,
) (map[string]any, error) {
	// Get the file path from the response
	filePath, ok := response["path"].(string)
	if !ok {
		return response, nil // Not a string path, return unchanged
	}

	// Check if it's an image file - convert common image formats
	ext := strings.ToLower(filepath.Ext(filePath))
	imageExtensions := []string{
		".jpg", ".jpeg", ".png", ".bmp", ".gif", ".tiff", ".webp",
	}

	isImageFile := slices.Contains(imageExtensions, ext)

	if !isImageFile {
		// Not an image file, return original response unchanged
		return response, nil
	}

	// Convert the image file
	convertedPath, err := s.convertImageToYUV(
		filePath,
		logrus.WithField("file", filePath),
	)
	if err != nil {
		logrus.WithError(err).
			WithField("file", filePath).
			Error("Image conversion failed")

		return response, ctxerrors.Wrapf(err, "image conversion failed")
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
	}).Info("Image file converted successfully to YUV format")

	return newResponse, nil
}

// convertImageToYUV converts uploaded image files to YUV format for SPECTRUMPAINT
// using ImageMagick convert command.
// Returns: convertedPath, error.
func (s *PIrateRF) convertImageToYUV(
	inputPath string,
	logger *logrus.Entry,
) (string, error) {
	if err := s.ensureFilesDirsExist(); err != nil {
		return "", err
	}

	outputPath := s.getImageOutputPath(inputPath)

	// Handle .Y files that may just need moving
	if strings.HasSuffix(inputPath, ".Y") {
		return s.handleYFile(inputPath, outputPath)
	}

	// Convert regular image to .Y format
	return s.convertImageToYFormat(inputPath, outputPath, logger)
}

func (s *PIrateRF) getImageOutputPath(inputPath string) string {
	imagesUploadsDir := path.Join(s.config.FilesDir, imagesUploadsPath)
	baseFilename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	return filepath.Join(imagesUploadsDir, baseFilename+".Y")
}

func (s *PIrateRF) handleYFile(inputPath, outputPath string) (string, error) {
	imagesUploadsDir := path.Join(s.config.FilesDir, imagesUploadsPath)

	if filepath.Dir(inputPath) == imagesUploadsDir {
		return inputPath, nil
	}

	if err := moveFile(inputPath, outputPath); err != nil {
		return "", ctxerrors.Wrapf(err, "failed to move .Y file")
	}

	return outputPath, nil
}

func (s *PIrateRF) convertImageToYFormat(inputPath, outputPath string, logger *logrus.Entry) (string, error) {
	tmpDir, tmpBasePath, err := s.createTempConversionDir()
	if err != nil {
		return "", err
	}
	defer s.cleanupTempDir(tmpDir, logger)

	if err := s.runImageMagickConversion(inputPath, tmpBasePath); err != nil {
		return "", err
	}

	if err := s.moveConvertedYFile(tmpBasePath, outputPath); err != nil {
		return "", err
	}

	s.cleanupOriginalFile(inputPath, logger)
	s.logConversionSuccess(inputPath, outputPath, logger)

	return outputPath, nil
}

func (s *PIrateRF) createTempConversionDir() (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "piraterf_image_convert_")
	if err != nil {
		return "", "", ctxerrors.Wrapf(err, "failed to create temp directory")
	}

	tmpBasePath := filepath.Join(tmpDir, "converted")

	return tmpDir, tmpBasePath, nil
}

func (s *PIrateRF) cleanupTempDir(tmpDir string, logger *logrus.Entry) {
	if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
		logger.WithError(removeErr).WithField("tmpDir", tmpDir).Warn("Failed to remove temporary directory")
	}
}

func (s *PIrateRF) runImageMagickConversion(inputPath, tmpBasePath string) error {
	ctx, cancel := context.WithTimeout(s.serviceCtx, audioConversionTimeout)
	defer cancel()

	process, err := s.commander.Start(ctx, "convert", []string{
		inputPath,
		"-resize", "320x",
		"-flip",
		"-quantize", "YUV",
		"-dither", "FloydSteinberg",
		"-colors", "4",
		"-interlace", "partition",
		tmpBasePath + ".yuv",
	})
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to start convert command")
	}

	if err := process.Wait(); err != nil {
		return ctxerrors.Wrapf(err, "convert command failed")
	}

	return nil
}

func (s *PIrateRF) moveConvertedYFile(tmpBasePath, outputPath string) error {
	tmpYPath := tmpBasePath + ".Y"
	if _, err := os.Stat(tmpYPath); err != nil {
		return ctxerrors.Wrapf(err, "converted .Y file not found")
	}

	if err := moveFile(tmpYPath, outputPath); err != nil {
		return ctxerrors.Wrapf(err, "failed to move converted .Y file")
	}

	return nil
}

func (s *PIrateRF) cleanupOriginalFile(inputPath string, logger *logrus.Entry) {
	if removeErr := os.Remove(inputPath); removeErr != nil {
		logger.WithError(removeErr).WithField("file", inputPath).Warn("Failed to remove original image file")
	}
}

func (s *PIrateRF) logConversionSuccess(inputPath, outputPath string, logger *logrus.Entry) {
	logger.WithFields(logrus.Fields{
		"original":  inputPath,
		"converted": outputPath,
	}).Info("Image converted to YUV format for SPECTRUMPAINT")
}

// moveFile moves a file from src to dst, handling cross-device link errors.
func moveFile(src, dst string) error {
	// Try rename first (fastest if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename failed, try copy + delete
	if err := copyFileStream(src, dst); err != nil {
		return ctxerrors.Wrapf(err, "failed to copy file")
	}

	// Remove source file after successful copy
	if err := os.Remove(src); err != nil {
		return ctxerrors.Wrapf(err, "failed to remove source file after copy")
	}

	return nil
}

// copyFileStream copies a file from src to dst using io.Copy.
func copyFileStream(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to open source file")
	}

	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			logrus.WithError(closeErr).Warn("Failed to close source file")
		}
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to create destination file")
	}

	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			logrus.WithError(closeErr).Warn("Failed to close destination file")
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to copy file content")
	}

	return nil
}
