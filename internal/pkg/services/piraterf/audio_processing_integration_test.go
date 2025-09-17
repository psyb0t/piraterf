package piraterf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/constants"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePlaylistFromFiles_Integration(t *testing.T) {
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
		filepath.Join(fixturesDir, "test_2s.mp3"), // 2 seconds
		filepath.Join(fixturesDir, "test_3s.wav"), // 3 seconds
		filepath.Join(fixturesDir, "test_4s.wav"), // 4 seconds
	}

	// Verify all fixture files exist
	for _, filePath := range testFiles {
		require.FileExists(t, filePath, "Fixture file should exist: %s", filePath)
	}

	// Test case: Create playlist from multiple files
	t.Run("CreatePlaylistFromMultipleFiles", func(t *testing.T) {
		playlistName := "test_playlist.wav"

		// Call the function under test
		outputPath, err := service.createPlaylistFromFiles(playlistName, testFiles)
		require.NoError(t, err, "createPlaylistFromFiles should not return error")
		require.NotEmpty(t, outputPath, "Output path should not be empty")

		// Verify the output file exists
		require.FileExists(t, outputPath, "Output playlist file should exist")

		// Verify the output file is in the expected directory structure
		expectedDir := filepath.Join(tempDir, audioFilesDir, uploadsSubdir)
		assert.True(t, strings.HasPrefix(outputPath, expectedDir),
			"Output path should be in uploads directory")

		// Use sox directly to verify the output format and duration
		cmd := commander.New()
		stdout, stderr, err := cmd.Output(context.Background(), "sox", []string{"--info", outputPath})
		require.NoError(t, err, "Should be able to get sox info, stderr: %s", string(stderr))

		output := string(stdout)
		t.Logf("Sox output for created playlist:\n%s", output)

		// Verify key properties are present in the sox output
		assert.Contains(t, output, "Sample Rate    : 48000", "Should have 48kHz sample rate")
		assert.Contains(t, output, "Channels       : 1", "Should be mono (1 channel)")
		assert.Contains(t, output, "Precision      : 16-bit", "Should be 16-bit")
		assert.Contains(t, output, "Sample Encoding: 16-bit Signed Integer PCM", "Should be PCM encoding")

		// Verify duration is approximately the sum of input files (2s + 3s + 4s = 9s)
		// Look for "Duration       : 00:00:09" pattern
		assert.Contains(t, output, "Duration       : 00:00:0", "Duration should start with 00:00:0")

		// More flexible duration check - should be between 8.5 and 9.5 seconds
		if strings.Contains(output, "Duration") {
			// Extract duration line for more detailed logging
			lines := strings.SplitSeq(output, "\n")
			for line := range lines {
				if strings.Contains(line, "Duration") {
					t.Logf("Duration line: %s", strings.TrimSpace(line))
					// Should contain approximately 9 seconds worth of content
					assert.True(t,
						strings.Contains(line, "00:00:08") ||
							strings.Contains(line, "00:00:09") ||
							strings.Contains(line, "00:00:10"),
						"Duration should be approximately 9 seconds, got: %s", line)

					break
				}
			}
		}

		// Log success info
		t.Logf("Successfully created playlist at: %s", outputPath)
	})

	// Test case: Create playlist with single file
	t.Run("CreatePlaylistFromSingleFile", func(t *testing.T) {
		playlistName := "single_file_playlist"
		singleFile := []string{testFiles[0]} // Just the 2-second MP3

		outputPath, err := service.createPlaylistFromFiles(playlistName, singleFile)
		require.NoError(t, err, "createPlaylistFromFiles should not return error")
		require.FileExists(t, outputPath, "Output playlist file should exist")

		// Verify it has .wav extension
		assert.True(t, strings.HasSuffix(outputPath, constants.FileExtensionWAV),
			"Output file should have .wav extension")

		// Use sox to verify the output
		cmd := commander.New()
		stdout, stderr, err := cmd.Output(context.Background(), "sox", []string{"--info", outputPath})
		require.NoError(t, err, "Should be able to get sox info, stderr: %s", string(stderr))

		output := string(stdout)
		t.Logf("Sox output for single file playlist:\n%s", output)

		// Verify format
		assert.Contains(t, output, "Sample Rate    : 48000", "Should have 48kHz sample rate")
		assert.Contains(t, output, "Channels       : 1", "Should be mono")
		assert.Contains(t, output, "Precision      : 16-bit", "Should be 16-bit")

		// Duration should be approximately 2 seconds
		assert.Contains(t, output, "Duration       : 00:00:02", "Duration should be approximately 2 seconds")
	})

	// Test case: Empty files list should return error
	t.Run("EmptyFilesList", func(t *testing.T) {
		playlistName := "empty_playlist.wav"
		emptyFiles := []string{}

		_, err := service.createPlaylistFromFiles(playlistName, emptyFiles)
		assert.Error(t, err, "Should return error for empty files list")
	})

	// Test case: Non-existent file should return error
	t.Run("NonExistentFile", func(t *testing.T) {
		playlistName := "nonexistent_playlist.wav"
		nonExistentFiles := []string{"/path/that/does/not/exist.wav"}

		_, err := service.createPlaylistFromFiles(playlistName, nonExistentFiles)
		assert.Error(t, err, "Should return error for non-existent file")
	})

	// Test case: Create playlist with custom output directory
	t.Run("CreatePlaylistWithCustomDirectory", func(t *testing.T) {
		customDir := t.TempDir() // Create temporary directory for test
		playlistName := "custom_dir_playlist.wav"

		// Use first two test files
		testFilesSubset := testFiles[:2]

		// Call with custom directory
		outputPath, err := service.createPlaylistFromFiles(playlistName, testFilesSubset, customDir)
		require.NoError(t, err, "createPlaylistFromFiles with custom dir should not return error")
		require.NotEmpty(t, outputPath, "Output path should not be empty")

		// Verify the output file exists
		require.FileExists(t, outputPath, "Output playlist file should exist")

		// Verify the output file is in the custom directory
		expectedPath := filepath.Join(customDir, playlistName)
		assert.Equal(t, expectedPath, outputPath, "Output path should be in custom directory")

		// Verify the file is actually in the custom directory
		assert.True(t, strings.HasPrefix(outputPath, customDir),
			"Output path should start with custom directory path")

		// Use sox to verify the output format
		cmd := commander.New()
		stdout, stderr, err := cmd.Output(context.Background(), "sox", []string{"--info", outputPath})
		require.NoError(t, err, "Should be able to get sox info, stderr: %s", string(stderr))

		output := string(stdout)
		t.Logf("Sox output for custom directory playlist:\n%s", output)

		// Verify format properties
		assert.Contains(t, output, "Sample Rate    : 48000", "Should have 48kHz sample rate")
		assert.Contains(t, output, "Channels       : 1", "Should be mono (1 channel)")
		assert.Contains(t, output, "Precision      : 16-bit", "Should be 16-bit")
		assert.Contains(t, output, "Sample Encoding: 16-bit Signed Integer PCM", "Should be PCM encoding")

		// Verify duration is approximately the sum of first two files (2s + 3s = 5s)
		if strings.Contains(output, "Duration") {
			lines := strings.SplitSeq(output, "\n")
			for line := range lines {
				if strings.Contains(line, "Duration") {
					t.Logf("Duration line: %s", strings.TrimSpace(line))
					assert.True(t,
						strings.Contains(line, "00:00:04") ||
							strings.Contains(line, "00:00:05") ||
							strings.Contains(line, "00:00:06"),
						"Duration should be approximately 5 seconds, got: %s", line)

					break
				}
			}
		}

		t.Logf("Successfully created playlist in custom directory: %s", outputPath)
	})

	// Test case: Create playlist in /tmp directory (temporary playlist)
	t.Run("CreateTemporaryPlaylist", func(t *testing.T) {
		playlistName := "temp_playlist" // No .wav extension to test auto-append
		tmpDir := "/tmp"

		// Use single test file
		singleFile := []string{testFiles[0]}

		// Call with /tmp directory
		outputPath, err := service.createPlaylistFromFiles(playlistName, singleFile, tmpDir)
		require.NoError(t, err, "createPlaylistFromFiles with /tmp should not return error")
		require.NotEmpty(t, outputPath, "Output path should not be empty")

		// Verify the output file exists
		require.FileExists(t, outputPath, "Output playlist file should exist")

		// Verify the output file is in /tmp
		assert.True(t, strings.HasPrefix(outputPath, "/tmp/"),
			"Output path should be in /tmp directory")

		// Verify .wav extension was auto-appended
		assert.True(t, strings.HasSuffix(outputPath, constants.FileExtensionWAV),
			"Output file should have .wav extension")

		expectedPath := filepath.Join(tmpDir, playlistName+constants.FileExtensionWAV)
		assert.Equal(t, expectedPath, outputPath, "Output path should match expected temp path")

		// Use sox to verify the output format
		cmd := commander.New()
		stdout, stderr, err := cmd.Output(context.Background(), "sox", []string{"--info", outputPath})
		require.NoError(t, err, "Should be able to get sox info, stderr: %s", string(stderr))

		output := string(stdout)
		t.Logf("Sox output for temporary playlist:\n%s", output)

		// Verify format properties
		assert.Contains(t, output, "Sample Rate    : 48000", "Should have 48kHz sample rate")
		assert.Contains(t, output, "Channels       : 1", "Should be mono")
		assert.Contains(t, output, "Precision      : 16-bit", "Should be 16-bit")

		// Duration should be approximately 2 seconds (single file)
		assert.Contains(t, output, "Duration       : 00:00:02", "Duration should be approximately 2 seconds")

		// Clean up the temporary file
		err = os.Remove(outputPath)
		require.NoError(t, err, "Should be able to clean up temporary file")

		t.Logf("Successfully created and cleaned up temporary playlist: %s", outputPath)
	})
}
