package piraterf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/psyb0t/commander"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertImageToYUV_Integration(t *testing.T) {
	// Set up logger to debug level for tests
	logrus.SetLevel(logrus.DebugLevel)

	// Create temporary directory for test output
	tempDir := t.TempDir()

	// Create PIrateRF instance with test config
	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
		serviceCtx: context.Background(), // Set context for tests
	}

	// Setup fixture files paths - assuming they exist in project root
	projectRoot := getProjectRoot(t)
	fixturesDir := filepath.Join(projectRoot, ".fixtures")

	// Define test fixture files
	testFiles := []string{
		filepath.Join(fixturesDir, "test_red_100x50.png"),     // PNG image
		filepath.Join(fixturesDir, "test_gradient_200x100.jpg"), // JPEG image
	}

	// Verify test fixture files exist
	for _, file := range testFiles {
		require.FileExists(t, file, "Test fixture file should exist: %s", file)
	}

	t.Run("ConvertImageToYUV_PNG", func(t *testing.T) {
		// Copy test file to temp upload location (simulating file upload)
		uploadPath := filepath.Join(tempDir, "uploaded_image.png")
		err := copyFile(testFiles[0], uploadPath)
		require.NoError(t, err, "Should be able to copy test file")

		// Convert the image to YUV
		logger := logrus.WithField("test", "ConvertImageToYUV_PNG")
		outputPath, err := service.convertImageToYUV(uploadPath, logger)
		require.NoError(t, err, "Should convert PNG image to YUV without error")

		// Verify the output file exists and has correct extension
		assert.FileExists(t, outputPath, "Converted YUV file should exist")
		assert.True(t, strings.HasSuffix(outputPath, ".Y"), "Output file should have .Y extension")

		// Verify the file is in the correct directory
		expectedDir := filepath.Join(tempDir, imagesUploadsPath)
		actualDir := filepath.Dir(outputPath)
		assert.Equal(t, expectedDir, actualDir, "Output should be in images uploads directory")

		// Verify the original file was cleaned up
		assert.NoFileExists(t, uploadPath, "Original uploaded file should be removed")

		// Verify the output file has content (non-zero size)
		stat, err := os.Stat(outputPath)
		require.NoError(t, err, "Should be able to stat output file")
		assert.Greater(t, stat.Size(), int64(0), "Output file should have content")

		t.Logf("Successfully converted PNG image to YUV: %s", outputPath)
	})

	t.Run("ConvertImageToYUV_JPEG", func(t *testing.T) {
		// Copy test file to temp upload location (simulating file upload)
		uploadPath := filepath.Join(tempDir, "uploaded_image.jpg")
		err := copyFile(testFiles[1], uploadPath)
		require.NoError(t, err, "Should be able to copy test file")

		// Convert the image to YUV
		logger := logrus.WithField("test", "ConvertImageToYUV_JPEG")
		outputPath, err := service.convertImageToYUV(uploadPath, logger)
		require.NoError(t, err, "Should convert JPEG image to YUV without error")

		// Verify the output file exists and has correct extension
		assert.FileExists(t, outputPath, "Converted YUV file should exist")
		assert.True(t, strings.HasSuffix(outputPath, ".Y"), "Output file should have .Y extension")

		// Verify the file is in the correct directory
		expectedDir := filepath.Join(tempDir, imagesUploadsPath)
		actualDir := filepath.Dir(outputPath)
		assert.Equal(t, expectedDir, actualDir, "Output should be in images uploads directory")

		// Verify the original file was cleaned up
		assert.NoFileExists(t, uploadPath, "Original uploaded file should be removed")

		// Verify the output file has content (non-zero size)
		stat, err := os.Stat(outputPath)
		require.NoError(t, err, "Should be able to stat output file")
		assert.Greater(t, stat.Size(), int64(0), "Output file should have content")

		t.Logf("Successfully converted JPEG image to YUV: %s", outputPath)
	})

	t.Run("ConvertImageToYUV_AlreadyYFormat", func(t *testing.T) {
		// Create a fake .Y file in upload directory
		service.ensureFilesDirsExist()
		yFilePath := filepath.Join(tempDir, imagesUploadsPath, "existing.Y")
		err := os.WriteFile(yFilePath, []byte("fake YUV data"), 0644)
		require.NoError(t, err, "Should be able to create test .Y file")

		// Try to convert it (should return the same path without conversion)
		logger := logrus.WithField("test", "ConvertImageToYUV_AlreadyYFormat")
		outputPath, err := service.convertImageToYUV(yFilePath, logger)
		require.NoError(t, err, "Should handle .Y file without error")

		// Should return the same path
		assert.Equal(t, yFilePath, outputPath, "Should return same path for .Y file")

		// File should still exist
		assert.FileExists(t, yFilePath, ".Y file should still exist")

		t.Logf("Correctly handled existing .Y file: %s", outputPath)
	})

	t.Run("ConvertImageToYUV_ExternalYFile", func(t *testing.T) {
		// Create a .Y file outside the target directory
		externalYPath := filepath.Join(tempDir, "external.Y")
		err := os.WriteFile(externalYPath, []byte("fake YUV data"), 0644)
		require.NoError(t, err, "Should be able to create external .Y file")

		// Convert it (should move to correct directory)
		logger := logrus.WithField("test", "ConvertImageToYUV_ExternalYFile")
		outputPath, err := service.convertImageToYUV(externalYPath, logger)
		require.NoError(t, err, "Should move external .Y file without error")

		// Should be in the correct directory
		expectedDir := filepath.Join(tempDir, imagesUploadsPath)
		actualDir := filepath.Dir(outputPath)
		assert.Equal(t, expectedDir, actualDir, "Output should be in images uploads directory")

		// Original file should be moved (no longer exist)
		assert.NoFileExists(t, externalYPath, "Original external .Y file should be moved")

		// New file should exist
		assert.FileExists(t, outputPath, "Moved .Y file should exist")

		t.Logf("Successfully moved external .Y file: %s", outputPath)
	})
}

func TestImageConversionPostprocessor_Integration(t *testing.T) {
	// Set up logger to debug level for tests
	logrus.SetLevel(logrus.DebugLevel)

	// Create temporary directory for test output
	tempDir := t.TempDir()

	// Create PIrateRF instance with test config
	service := &PIrateRF{
		config: Config{
			FilesDir: tempDir,
		},
		serviceCtx: context.Background(),
	}

	// Setup fixture files paths
	projectRoot := getProjectRoot(t)
	fixturesDir := filepath.Join(projectRoot, ".fixtures")
	testImagePath := filepath.Join(fixturesDir, "test_red_100x50.png")

	// Verify test fixture exists
	require.FileExists(t, testImagePath, "Test fixture file should exist")

	t.Run("ImageConversionPostprocessor_PNG", func(t *testing.T) {
		// Copy test file to temp upload location
		uploadPath := filepath.Join(tempDir, "uploaded.png")
		err := copyFile(testImagePath, uploadPath)
		require.NoError(t, err, "Should be able to copy test file")

		// Create response like the file upload handler would
		response := map[string]any{
			"path":           uploadPath,
			"saved_filename": "uploaded.png",
			"size":           int64(280), // approximate size of our test PNG
		}

		// Process the response
		newResponse, err := service.imageConversionPostprocessor(response)
		require.NoError(t, err, "Should process image without error")

		// Verify the response was modified
		assert.True(t, newResponse["converted"].(bool), "Response should indicate conversion happened")

		convertedPath := newResponse["path"].(string)
		assert.True(t, strings.HasSuffix(convertedPath, ".Y"), "Converted path should have .Y extension")

		convertedFilename := newResponse["saved_filename"].(string)
		assert.True(t, strings.HasSuffix(convertedFilename, ".Y"), "Converted filename should have .Y extension")

		// Verify the converted file exists
		assert.FileExists(t, convertedPath, "Converted file should exist")

		// Verify it's in the correct directory
		expectedDir := filepath.Join(tempDir, imagesUploadsPath)
		actualDir := filepath.Dir(convertedPath)
		assert.Equal(t, expectedDir, actualDir, "Converted file should be in images uploads directory")

		t.Logf("Successfully processed image upload and converted to: %s", convertedPath)
	})

	t.Run("ImageConversionPostprocessor_NonImage", func(t *testing.T) {
		// Create a non-image file
		textPath := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(textPath, []byte("not an image"), 0644)
		require.NoError(t, err, "Should be able to create test text file")

		// Create response for non-image file
		response := map[string]any{
			"path":           textPath,
			"saved_filename": "test.txt",
			"size":           int64(13),
		}

		// Process the response
		newResponse, err := service.imageConversionPostprocessor(response)
		require.NoError(t, err, "Should process non-image without error")

		// Verify the response was NOT modified (no conversion for non-images)
		assert.Equal(t, response, newResponse, "Response should be unchanged for non-image files")

		// Original file should still exist
		assert.FileExists(t, textPath, "Non-image file should not be removed")

		t.Logf("Correctly ignored non-image file: %s", textPath)
	})
}

func TestImageProcessingWithImageMagick_Integration(t *testing.T) {
	// Verify ImageMagick convert command is available
	cmd := commander.New()

	stdout, stderr, err := cmd.Output(context.Background(), "convert", []string{"-version"})
	if err != nil {
		t.Skipf("ImageMagick convert command not available: %v, stderr: %s", err, string(stderr))
	}

	output := string(stdout)
	assert.Contains(t, output, "ImageMagick", "Should be ImageMagick convert command")

	t.Logf("ImageMagick version info: %s", strings.Split(output, "\n")[0])
}