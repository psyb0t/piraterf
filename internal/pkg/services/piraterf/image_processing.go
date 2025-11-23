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

	"github.com/psyb0t/common-go/constants"
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
		constants.FileExtensionJPG,
		constants.FileExtensionJPEG,
		constants.FileExtensionPNG,
		constants.FileExtensionBMP,
		constants.FileExtensionGIF,
		constants.FileExtensionTIFF,
		constants.FileExtensionWEBP,
	}

	isImageFile := slices.Contains(imageExtensions, ext)

	if !isImageFile {
		// Not an image file, return original response unchanged
		return response, nil
	}

	// Convert the image file to both YUV and RGB formats
	convertedPath, rgbPath, err := s.convertImageToFormats(
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
		"original":      filePath,
		"converted_y":   convertedPath,
		"converted_rgb": rgbPath,
	}).Info("Image file converted successfully to YUV and RGB formats")

	return newResponse, nil
}

// convertImageToFormats converts uploaded image files to both
// YUV and RGB formats using ImageMagick convert command.
// Returns: yuvPath, rgbPath, error.
func (s *PIrateRF) convertImageToFormats(
	inputPath string,
	logger *logrus.Entry,
) (string, string, error) {
	// Create a copy of the input file for RGB conversion since YUV
	// will delete the original
	tempInputPath := inputPath + ".temp"
	if err := s.copyFileForTemp(inputPath, tempInputPath); err != nil {
		return "", "", ctxerrors.Wrapf(err, "failed to create temp copy")
	}

	yuvPath, err := s.convertImageToYUV(inputPath, logger)
	if err != nil {
		if removeErr := os.Remove(tempInputPath); removeErr != nil {
			logger.WithError(removeErr).Warn("Failed to cleanup temp file")
		}

		return "", "", err
	}

	rgbOutputPath := s.getImageRGBOutputPath(inputPath)

	rgbPath, err := s.convertImageToRGBFormat(
		tempInputPath,
		rgbOutputPath,
		logger,
	)
	if err != nil {
		return "", "", err
	}

	return yuvPath, rgbPath, nil
}

// convertImageToYUV converts uploaded image files to YUV format
// for SPECTRUMPAINT using ImageMagick convert command.
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
	base := filepath.Base(inputPath)
	baseFilename := strings.TrimSuffix(base, filepath.Ext(inputPath))

	return filepath.Join(imagesUploadsDir, baseFilename+".Y")
}

func (s *PIrateRF) getImageRGBOutputPath(inputPath string) string {
	imagesUploadsDir := path.Join(s.config.FilesDir, imagesUploadsPath)
	base := filepath.Base(inputPath)
	baseFilename := strings.TrimSuffix(base, filepath.Ext(inputPath))

	return filepath.Join(imagesUploadsDir, baseFilename+".rgb")
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

func (s *PIrateRF) convertImageToYFormat(
	inputPath, outputPath string,
	logger *logrus.Entry,
) (string, error) {
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

func (s *PIrateRF) convertImageToRGBFormat(
	inputPath, outputPath string,
	logger *logrus.Entry,
) (string, error) {
	ctx, cancel := context.WithTimeout(s.serviceCtx, audioConversionTimeout)
	defer cancel()

	// Convert image to 320x256 RGB format for PISSTV
	process, err := s.commander.Start(ctx, "convert", []string{
		inputPath,
		"-resize", "320x256!", // Force exact 320x256 dimensions
		"-depth", "8", // 8 bits per channel
		"rgb:" + outputPath, // Output as raw RGB format
	})
	if err != nil {
		return "", ctxerrors.Wrapf(err, "failed to start convert command for RGB")
	}

	if err := process.Wait(); err != nil {
		return "", ctxerrors.Wrapf(err, "convert command failed for RGB")
	}

	s.cleanupOriginalFile(inputPath, logger)
	logger.WithFields(logrus.Fields{
		"original":  inputPath,
		"converted": outputPath,
	}).Info("Image converted to RGB format for PISSTV")

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

func (s *PIrateRF) cleanupTempDir(
	tmpDir string,
	logger *logrus.Entry,
) {
	if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
		logger.
			WithError(removeErr).
			WithField("tmpDir", tmpDir).
			Warn("Failed to remove temporary directory")
	}
}

func (s *PIrateRF) runImageMagickConversion(
	inputPath, tmpBasePath string,
) error {
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

func (s *PIrateRF) moveConvertedYFile(
	tmpBasePath, outputPath string,
) error {
	tmpYPath := tmpBasePath + ".Y"
	if _, err := os.Stat(tmpYPath); err != nil {
		return ctxerrors.Wrapf(err, "converted .Y file not found")
	}

	if err := moveFile(tmpYPath, outputPath); err != nil {
		return ctxerrors.Wrapf(err, "failed to move converted .Y file")
	}

	return nil
}

func (s *PIrateRF) cleanupOriginalFile(
	inputPath string,
	logger *logrus.Entry,
) {
	if removeErr := os.Remove(inputPath); removeErr != nil {
		logger.
			WithError(removeErr).
			WithField("file", inputPath).
			Warn("Failed to remove original image file")
	}
}

func (s *PIrateRF) logConversionSuccess(
	inputPath, outputPath string,
	logger *logrus.Entry,
) {
	logger.WithFields(logrus.Fields{
		"original":  inputPath,
		"converted": outputPath,
	}).Info("Image converted to YUV")
}

// copyFileForTemp copies a file from src to dst using io.Copy
// for temporary file creation.
func (s *PIrateRF) copyFileForTemp(src, dst string) error {
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
