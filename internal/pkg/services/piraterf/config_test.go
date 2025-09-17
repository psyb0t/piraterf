package piraterf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(t *testing.T, cfg Config)
	}{
		{
			name:        "default config values",
			envVars:     map[string]string{},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, defaultHTMLDir, cfg.HTMLDir)
				assert.Equal(t, defaultStaticDir, cfg.StaticDir)
				assert.Equal(t, defaultFilesDir, cfg.FilesDir)
				assert.Equal(t, defaultUploadDir, cfg.UploadDir)
			},
		},
		{
			name: "custom config from environment variables",
			envVars: map[string]string{
				envVarNameHTMLDir:          "/custom/html",
				envVarNameStaticDir:        "/custom/static",
				envVarNamePiraterfFilesDir: "/custom/files",
				envVarNameUploadDir:        "/custom/uploads",
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "/custom/html", cfg.HTMLDir)
				assert.Equal(t, "/custom/static", cfg.StaticDir)
				assert.Equal(t, "/custom/files", cfg.FilesDir)
				assert.Equal(t, "/custom/uploads", cfg.UploadDir)
			},
		},
		{
			name: "partial custom config",
			envVars: map[string]string{
				envVarNameHTMLDir:          "/custom/html",
				envVarNamePiraterfFilesDir: "/custom/files",
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "/custom/html", cfg.HTMLDir)
				assert.Equal(t, defaultStaticDir, cfg.StaticDir)
				assert.Equal(t, "/custom/files", cfg.FilesDir)
				assert.Equal(t, defaultUploadDir, cfg.UploadDir)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := parseConfig()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}