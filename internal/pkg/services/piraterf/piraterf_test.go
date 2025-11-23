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
				t.Helper()
				t.Setenv("ENV", "dev")

				// Create required directories
				require.NoError(t, os.MkdirAll("static", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("static"); err != nil {
						t.Logf("Failed to remove static dir: %v", err)
					}
				})

				require.NoError(t, os.MkdirAll("files", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("files"); err != nil {
						t.Logf("Failed to remove files dir: %v", err)
					}
				})

				require.NoError(t, os.MkdirAll("uploads", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("uploads"); err != nil {
						t.Logf("Failed to remove uploads dir: %v", err)
					}
				})
			},
			expectError: false,
			validate: func(t *testing.T, service *PIrateRF) {
				t.Helper()
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
				if err := service.Stop(ctx); err != nil {
					t.Logf("Failed to stop service: %v", err)
				}
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
				t.Helper()
				t.Setenv("ENV", "dev")

				// Create required directories with custom paths
				require.NoError(t, os.MkdirAll("test_static", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("test_static"); err != nil {
						t.Logf("Failed to remove test_static: %v", err)
					}
				})

				require.NoError(t, os.MkdirAll("test_files", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("test_files"); err != nil {
						t.Logf("Failed to remove test_files: %v", err)
					}
				})

				require.NoError(t, os.MkdirAll("test_uploads", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("test_uploads"); err != nil {
						t.Logf("Failed to remove test_uploads: %v", err)
					}
				})
			},
			expectError: false,
			validate: func(t *testing.T, service *PIrateRF) {
				t.Helper()
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
				if err := service.Stop(ctx); err != nil {
					t.Logf("Failed to stop service: %v", err)
				}
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
				t.Helper()
				// Ensure directory doesn't exist
				_ = os.RemoveAll("./test_upload_dir")
				t.Cleanup(func() {
					if err := os.RemoveAll("./test_upload_dir"); err != nil {
						t.Logf("Failed to remove test_upload_dir: %v", err)
					}
				})
			},
			expectError: false,
		},
		{
			name: "upload directory already exists",
			config: Config{
				UploadDir: "./existing_upload_dir",
			},
			setupFunc: func(t *testing.T) {
				t.Helper()
				// Create directory first
				require.NoError(t, os.MkdirAll("./existing_upload_dir", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("./existing_upload_dir"); err != nil {
						t.Logf("Failed to remove existing_upload_dir: %v", err)
					}
				})
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
				t.Helper()
				// Ensure directories don't exist
				_ = os.RemoveAll("./test_files_dir")
				t.Cleanup(func() {
					if err := os.RemoveAll("./test_files_dir"); err != nil {
						t.Logf("Failed to remove test_files_dir: %v", err)
					}
				})
			},
			expectError: false,
			validate: func(t *testing.T, config Config) {
				t.Helper()
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
				imagesUploadsDir := path.Join(
					config.FilesDir,
					imagesFilesDir,
					uploadsSubdir,
				)
				info, err = os.Stat(imagesUploadsDir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check preset directories for all modules
				rpitx := gorpitx.GetInstance()
				moduleNames := rpitx.GetSupportedModules()
				for _, moduleName := range moduleNames {
					modulePresetDir := path.Join(config.FilesDir, presetsDir, moduleName)
					info, err = os.Stat(modulePresetDir)
					assert.NoError(t, err)
					assert.True(t, info.IsDir())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			service := &PIrateRF{
				config: tt.config,
				rpitx:  gorpitx.GetInstance(),
			}
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
				t.Helper()
				// Create static directory
				require.NoError(t, os.MkdirAll("./test_static", 0o750))
				t.Cleanup(func() {
					if err := os.RemoveAll("./test_static"); err != nil {
						t.Logf("Failed to remove test_static: %v", err)
					}
				})
			},
			expectError: false,
			validate: func(t *testing.T, config Config) {
				t.Helper()
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

func setupTestService(t *testing.T) *PIrateRF {
	t.Helper()
	t.Setenv("ENV", "dev")

	// Create required directories
	require.NoError(t, os.MkdirAll("static", 0o750))
	t.Cleanup(func() {
		_ = os.RemoveAll("static")
	})

	require.NoError(t, os.MkdirAll("files", 0o750))
	t.Cleanup(func() {
		_ = os.RemoveAll("files")
	})

	require.NoError(t, os.MkdirAll("uploads", 0o750))
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	service, err := New()
	require.NoError(t, err)

	return service
}

func TestPIrateRF_Stop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "stop service successfully"},
		{name: "stop service multiple times (idempotent)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := setupTestService(t)
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
		runDuration time.Duration
		stopMethod  string // "context" or "service"
	}{
		{
			name:        "run and stop via context cancellation",
			runDuration: 50 * time.Millisecond,
			stopMethod:  "context",
		},
		{
			name:        "run and stop via service stop",
			runDuration: 50 * time.Millisecond,
			stopMethod:  "service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := setupTestService(t)
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

				return
			}

			assert.NoError(t, err)
		})
	}
}
