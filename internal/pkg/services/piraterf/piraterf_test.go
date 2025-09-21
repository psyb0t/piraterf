package piraterf

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/psyb0t/common-go/env"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T)
		expectError bool
		validate    func(t *testing.T, service *PIrateRF)
	}{
		{
			name: "successful service creation",
			setupFunc: func(t *testing.T) {
				t.Setenv("ENV", "dev")

				// Create required directories
				require.NoError(t, os.MkdirAll("static", 0o755))
				t.Cleanup(func() { os.RemoveAll("static") })

				require.NoError(t, os.MkdirAll("files", 0o755))
				t.Cleanup(func() { os.RemoveAll("files") })

				require.NoError(t, os.MkdirAll("uploads", 0o755))
				t.Cleanup(func() { os.RemoveAll("uploads") })
			},
			expectError: false,
			validate: func(t *testing.T, service *PIrateRF) {
				assert.NotNil(t, service)
				assert.Equal(t, ServiceName, service.Name())
				assert.NotNil(t, service.config)
				assert.NotNil(t, service.rpitx)
				assert.NotNil(t, service.httpServer)
				assert.NotNil(t, service.websocketHub)
				assert.NotNil(t, service.executionManager)
				assert.NotNil(t, service.doneCh)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			service, err := New()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)

				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, service)
			}

			// Cleanup
			if service != nil {
				ctx := context.Background()
				service.Stop(ctx)
			}
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		setupFunc   func(t *testing.T)
		expectError bool
		validate    func(t *testing.T, service *PIrateRF)
	}{
		{
			name: "successful service creation with custom config",
			config: Config{
				HTMLDir:   "./test_html",
				StaticDir: "./test_static",
				FilesDir:  "./test_files",
				UploadDir: "./test_uploads",
			},
			setupFunc: func(t *testing.T) {
				t.Setenv("ENV", "dev")

				// Create required directories with custom paths
				require.NoError(t, os.MkdirAll("test_static", 0o755))
				t.Cleanup(func() { os.RemoveAll("test_static") })

				require.NoError(t, os.MkdirAll("test_files", 0o755))
				t.Cleanup(func() { os.RemoveAll("test_files") })

				require.NoError(t, os.MkdirAll("test_uploads", 0o755))
				t.Cleanup(func() { os.RemoveAll("test_uploads") })
			},
			expectError: false,
			validate: func(t *testing.T, service *PIrateRF) {
				assert.NotNil(t, service)
				assert.Equal(t, "./test_html", service.config.HTMLDir)
				assert.Equal(t, "./test_static", service.config.StaticDir)
				assert.Equal(t, "./test_files", service.config.FilesDir)
				assert.Equal(t, "./test_uploads", service.config.UploadDir)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			service, err := NewWithConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)

				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, service)
			}

			// Cleanup
			if service != nil {
				ctx := context.Background()
				service.Stop(ctx)
			}
		})
	}
}

func TestPIrateRF_Name(t *testing.T) {
	service := &PIrateRF{}
	assert.Equal(t, ServiceName, service.Name())
}

func TestPIrateRF_ensureUploadDirExists(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		setupFunc   func(t *testing.T)
		expectError bool
	}{
		{
			name: "creates upload directory successfully",
			config: Config{
				UploadDir: "./test_upload_dir",
			},
			setupFunc: func(t *testing.T) {
				// Ensure directory doesn't exist
				os.RemoveAll("./test_upload_dir")
				t.Cleanup(func() { os.RemoveAll("./test_upload_dir") })
			},
			expectError: false,
		},
		{
			name: "upload directory already exists",
			config: Config{
				UploadDir: "./existing_upload_dir",
			},
			setupFunc: func(t *testing.T) {
				// Create directory first
				require.NoError(t, os.MkdirAll("./existing_upload_dir", 0o755))
				t.Cleanup(func() { os.RemoveAll("./existing_upload_dir") })
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			service := &PIrateRF{config: tt.config}
			err := service.ensureUploadDirExists()

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			// Verify directory exists
			info, err := os.Stat(tt.config.UploadDir)
			assert.NoError(t, err)
			assert.True(t, info.IsDir())
		})
	}
}

func TestPIrateRF_ensureFilesDirsExist(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		setupFunc   func(t *testing.T)
		expectError bool
		validate    func(t *testing.T, config Config)
	}{
		{
			name: "creates all required directories",
			config: Config{
				FilesDir: "./test_files_dir",
			},
			setupFunc: func(t *testing.T) {
				// Ensure directories don't exist
				os.RemoveAll("./test_files_dir")
				t.Cleanup(func() { os.RemoveAll("./test_files_dir") })
			},
			expectError: false,
			validate: func(t *testing.T, config Config) {
				// Check base files directory
				info, err := os.Stat(config.FilesDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check audio directory
				audioDir := path.Join(config.FilesDir, audioFilesDir)
				info, err = os.Stat(audioDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check audio uploads directory
				audioUploadsDir := path.Join(config.FilesDir, audioFilesDir, uploadsSubdir)
				info, err = os.Stat(audioUploadsDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check audio SFX directory
				audioSFXDirPath := path.Join(config.FilesDir, audioFilesDir, audioSFXDir)
				info, err = os.Stat(audioSFXDirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check images directory
				imagesDir := path.Join(config.FilesDir, imagesFilesDir)
				info, err = os.Stat(imagesDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check images uploads directory
				imagesUploadsDir := path.Join(config.FilesDir, imagesFilesDir, uploadsSubdir)
				info, err = os.Stat(imagesUploadsDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			service := &PIrateRF{config: tt.config}
			err := service.ensureFilesDirsExist()

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, tt.config)
			}
		})
	}
}

func TestPIrateRF_generateEnvJS(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		setupFunc   func(t *testing.T)
		expectError bool
		validate    func(t *testing.T, config Config)
	}{
		{
			name: "generates env.js file successfully",
			config: Config{
				StaticDir: "./test_static",
				FilesDir:  "./test_files",
			},
			setupFunc: func(t *testing.T) {
				// Create static directory
				require.NoError(t, os.MkdirAll("./test_static", 0o755))
				t.Cleanup(func() { os.RemoveAll("./test_static") })
			},
			expectError: false,
			validate: func(t *testing.T, config Config) {
				envJSPath := path.Join(config.StaticDir, envJSFilename)

				// Check file exists
				info, err := os.Stat(envJSPath)
				assert.NoError(t, err)
				assert.False(t, info.IsDir())

				// Check file content
				content, err := os.ReadFile(envJSPath)
				assert.NoError(t, err)
				assert.Contains(t, string(content), "window.PIrateRFConfig")
				assert.Contains(t, string(content), config.FilesDir)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			service := &PIrateRF{config: tt.config}
			err := service.generateEnvJS()

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, tt.config)
			}
		})
	}
}

func TestPIrateRF_Stop(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) *PIrateRF
	}{
		{
			name: "stop service successfully",
			setupFunc: func(t *testing.T) *PIrateRF {
				t.Setenv("ENV", "dev")

				// Create required directories
				require.NoError(t, os.MkdirAll("static", 0o755))
				t.Cleanup(func() { os.RemoveAll("static") })

				require.NoError(t, os.MkdirAll("files", 0o755))
				t.Cleanup(func() { os.RemoveAll("files") })

				require.NoError(t, os.MkdirAll("uploads", 0o755))
				t.Cleanup(func() { os.RemoveAll("uploads") })

				service, err := New()
				require.NoError(t, err)

				return service
			},
		},
		{
			name: "stop service multiple times (idempotent)",
			setupFunc: func(t *testing.T) *PIrateRF {
				t.Setenv("ENV", "dev")

				// Create required directories
				require.NoError(t, os.MkdirAll("static", 0o755))
				t.Cleanup(func() { os.RemoveAll("static") })

				require.NoError(t, os.MkdirAll("files", 0o755))
				t.Cleanup(func() { os.RemoveAll("files") })

				require.NoError(t, os.MkdirAll("uploads", 0o755))
				t.Cleanup(func() { os.RemoveAll("uploads") })

				service, err := New()
				require.NoError(t, err)

				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupFunc(t)
			require.NotNil(t, service)

			ctx := context.Background()

			// First stop should work
			err := service.Stop(ctx)
			assert.NoError(t, err)

			// Second stop should also work (idempotent)
			err = service.Stop(ctx)
			assert.NoError(t, err)
		})
	}
}

func TestPIrateRF_Run(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) *PIrateRF
		runDuration time.Duration
		stopMethod  string // "context" or "service"
	}{
		{
			name: "run and stop via context cancellation",
			setupFunc: func(t *testing.T) *PIrateRF {
				t.Setenv("ENV", "dev")

				// Create required directories
				require.NoError(t, os.MkdirAll("static", 0o755))
				t.Cleanup(func() { os.RemoveAll("static") })

				require.NoError(t, os.MkdirAll("files", 0o755))
				t.Cleanup(func() { os.RemoveAll("files") })

				require.NoError(t, os.MkdirAll("uploads", 0o755))
				t.Cleanup(func() { os.RemoveAll("uploads") })

				service, err := New()
				require.NoError(t, err)

				return service
			},
			runDuration: 50 * time.Millisecond,
			stopMethod:  "context",
		},
		{
			name: "run and stop via service stop",
			setupFunc: func(t *testing.T) *PIrateRF {
				t.Setenv("ENV", "dev")

				// Create required directories
				require.NoError(t, os.MkdirAll("static", 0o755))
				t.Cleanup(func() { os.RemoveAll("static") })

				require.NoError(t, os.MkdirAll("files", 0o755))
				t.Cleanup(func() { os.RemoveAll("files") })

				require.NoError(t, os.MkdirAll("uploads", 0o755))
				t.Cleanup(func() { os.RemoveAll("uploads") })

				service, err := New()
				require.NoError(t, err)

				return service
			},
			runDuration: 50 * time.Millisecond,
			stopMethod:  "service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupFunc(t)
			require.NotNil(t, service)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan error, 1)

			// Start service in goroutine
			go func() {
				done <- service.Run(ctx)
			}()

			// Let service start
			time.Sleep(tt.runDuration)

			// Stop using specified method
			switch tt.stopMethod {
			case "context":
				cancel()
			case "service":
				err := service.Stop(ctx)
				assert.NoError(t, err)
			}

			// Wait for service to complete
			select {
			case err := <-done:
				assert.NoError(t, err)
			case <-time.After(2 * time.Second):
				t.Fatal("service did not stop within timeout")
			}
		})
	}
}

func TestValidateModuleInDev(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name        string
		module      gorpitx.ModuleName
		expectError bool
	}{
		{
			name:        "supported module PIFMRDS",
			module:      gorpitx.ModuleNamePIFMRDS,
			expectError: false,
		},
		{
			name:        "supported module SPECTRUMPAINT",
			module:      gorpitx.ModuleNameSPECTRUMPAINT,
			expectError: false,
		},
		{
			name:        "supported module PICHIRP",
			module:      gorpitx.ModuleNamePICHIRP,
			expectError: false,
		},
		{
			name:        "supported module POCSAG",
			module:      gorpitx.ModuleNamePOCSAG,
			expectError: false,
		},
		{
			name:        "invalid module",
			module:      gorpitx.ModuleName("INVALID"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpitx := gorpitx.GetInstance()
			service := &PIrateRF{
				rpitx: rpitx,
			}

			logger := logrus.WithField("test", "validateModuleInDev")
			err := service.validateModuleInDev(tt.module, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
