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

	"github.com/psyb0t/commander"
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
	// Ensure images directory structure exists
	if err := s.ensureFilesDirsExist(); err != nil {
		return "", err
	}

	imagesUploadsDir := path.Join(s.config.FilesDir, imagesUploadsPath)

	// Generate output path in ./files/images/uploads with .Y extension
	baseFilename := strings.TrimSuffix(
		filepath.Base(inputPath),
		filepath.Ext(inputPath),
	)
	outputPath := filepath.Join(
		imagesUploadsDir,
		baseFilename+".Y",
	)

	// Check if the file is already in .Y format
	if strings.HasSuffix(inputPath, ".Y") {
		// If it's already a .Y file in the correct directory, no conversion needed
		if filepath.Dir(inputPath) == imagesUploadsDir {
			return inputPath, nil
		}

		// Move .Y file to correct directory
		if err := moveFile(inputPath, outputPath); err != nil {
			return "", ctxerrors.Wrapf(err, "failed to move .Y file")
		}

		return outputPath, nil
	}

	// Create temporary directory for ImageMagick conversion
	tmpDir, err := os.MkdirTemp("", "piraterf_image_convert_")
	if err != nil {
		return "", ctxerrors.Wrapf(err, "failed to create temp directory")
	}

	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			logger.WithError(removeErr).
				WithField("tmpDir", tmpDir).
				Warn("Failed to remove temporary directory")
		}
	}()

	// Generate temporary output base name for ImageMagick
	tmpBasePath := filepath.Join(tmpDir, "converted")

	// Use ImageMagick convert to create YUV format
	cmd := commander.New()

	ctx, cancel := context.WithTimeout(
		s.serviceCtx,
		audioConversionTimeout, // Reuse audio timeout for images
	)
	defer cancel()

	// convert input.jpg -resize 320x -flip -quantize YUV -dither FloydSteinberg
	// -colors 4 -interlace partition converted.yuv
	process, err := cmd.Start(ctx, "convert", []string{
		inputPath,
		"-resize", "320x", // Fixed width of 320 pixels
		"-flip",            // Flip image (required by rpitx)
		"-quantize", "YUV", // Convert to YUV color space
		"-dither", "FloydSteinberg", // Apply dithering
		"-colors", "4", // Reduce to 4 colors
		"-interlace", "partition", // Create separate Y, U, V files
		tmpBasePath + ".yuv",
	})
	if err != nil {
		return "", ctxerrors.Wrapf(err, "failed to start convert command")
	}

	// Wait for conversion to complete
	if err := process.Wait(); err != nil {
		return "", ctxerrors.Wrapf(err, "convert command failed")
	}

	// Move the .Y file (luminance channel) to final destination
	tmpYPath := tmpBasePath + ".Y"
	if _, err := os.Stat(tmpYPath); err != nil {
		return "", ctxerrors.Wrapf(err, "converted .Y file not found")
	}

	if err := moveFile(tmpYPath, outputPath); err != nil {
		return "", ctxerrors.Wrapf(err, "failed to move converted .Y file")
	}

	// Clean up the original uploaded image file
	if removeErr := os.Remove(inputPath); removeErr != nil {
		logger.WithError(removeErr).
			WithField("file", inputPath).
			Warn("Failed to remove original image file")
	}

	logger.WithFields(logrus.Fields{
		"original":  inputPath,
		"converted": outputPath,
	}).Info("Image converted to YUV format for SPECTRUMPAINT")

	return outputPath, nil
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
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to create destination file")
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to copy file content")
	}

	return nil
}
