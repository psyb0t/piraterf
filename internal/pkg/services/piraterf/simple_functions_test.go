package piraterf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
	}{
		{
			name:     "service name",
			constant: ServiceName,
			expected: "PIrateRF",
		},
		{
			name:     "audio files dir",
			constant: audioFilesDir,
			expected: "audio",
		},
		{
			name:     "uploads subdir",
			constant: uploadsSubdir,
			expected: "uploads",
		},
		{
			name:     "audio sfx dir",
			constant: audioSFXDir,
			expected: "sfx",
		},
		{
			name:     "images files dir",
			constant: imagesFilesDir,
			expected: "images",
		},
		{
			name:     "env js filename",
			constant: envJSFilename,
			expected: "env.js",
		},
		{
			name:     "audio sample rate",
			constant: audioSampleRate,
			expected: "48000",
		},
		{
			name:     "audio bit depth",
			constant: audioBitDepth,
			expected: "16",
		},
		{
			name:     "audio channels",
			constant: audioChannels,
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestConfigConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "html dir env var",
			constant: envVarNameHTMLDir,
			expected: "PIRATERF_HTMLDIR",
		},
		{
			name:     "static dir env var",
			constant: envVarNameStaticDir,
			expected: "PIRATERF_STATICDIR",
		},
		{
			name:     "files dir env var",
			constant: envVarNamePiraterfFilesDir,
			expected: "PIRATERF_FILESDIR",
		},
		{
			name:     "upload dir env var",
			constant: envVarNameUploadDir,
			expected: "PIRATERF_UPLOADDIR",
		},
		{
			name:     "default html dir",
			constant: defaultHTMLDir,
			expected: "./html",
		},
		{
			name:     "default static dir",
			constant: defaultStaticDir,
			expected: "./static",
		},
		{
			name:     "default files dir",
			constant: defaultFilesDir,
			expected: "./files",
		},
		{
			name:     "default upload dir",
			constant: defaultUploadDir,
			expected: "./uploads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestExecutionStateConstants(t *testing.T) {
	tests := []struct {
		name     string
		state    executionState
		expected executionState
	}{
		{
			name:     "idle state",
			state:    executionStateIdle,
			expected: 0,
		},
		{
			name:     "executing state",
			state:    executionStateExecuting,
			expected: 1,
		},
		{
			name:     "stopping state",
			state:    executionStateStopping,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state)
		})
	}
}

func TestPermissionsConstants(t *testing.T) {
	tests := []struct {
		name        string
		permissions interface{}
		expected    interface{}
	}{
		{
			name:        "directory permissions",
			permissions: dirPerms,
			expected:    0o750,
		},
		{
			name:        "file permissions",
			permissions: filePerms,
			expected:    0o600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.permissions)
		})
	}
}