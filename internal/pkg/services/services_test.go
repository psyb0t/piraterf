package services

import (
	"os"
	"testing"

	servicemanager "github.com/psyb0t/piraterf/internal/pkg/service-manager"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T)
		servicesEnv  string
		expectPanic  bool
		validate     func(t *testing.T)
	}{
		{
			name: "all services enabled by default",
			setupFunc: func(t *testing.T) {
				t.Setenv("ENV", "dev")

				// Create required directories
				assert.NoError(t, os.MkdirAll("static", 0755))
				t.Cleanup(func() { os.RemoveAll("static") })

				assert.NoError(t, os.MkdirAll("files", 0755))
				t.Cleanup(func() { os.RemoveAll("files") })

				assert.NoError(t, os.MkdirAll("uploads", 0755))
				t.Cleanup(func() { os.RemoveAll("uploads") })

				// Reset service manager
				servicemanager.ResetInstance()
			},
			servicesEnv: "",
			expectPanic: false,
			validate: func(t *testing.T) {
				// PIrateRF service should be added
				// We can't easily verify this without exposing internals,
				// but the function should complete without panic
			},
		},
		{
			name: "specific services enabled",
			setupFunc: func(t *testing.T) {
				t.Setenv("ENV", "dev")

				// Create required directories
				assert.NoError(t, os.MkdirAll("static", 0755))
				t.Cleanup(func() { os.RemoveAll("static") })

				assert.NoError(t, os.MkdirAll("files", 0755))
				t.Cleanup(func() { os.RemoveAll("files") })

				assert.NoError(t, os.MkdirAll("uploads", 0755))
				t.Cleanup(func() { os.RemoveAll("uploads") })

				// Reset service manager
				servicemanager.ResetInstance()
			},
			servicesEnv: "PIrateRF",
			expectPanic: false,
		},
		{
			name: "no services enabled",
			setupFunc: func(t *testing.T) {
				t.Setenv("ENV", "dev")

				// Reset service manager
				servicemanager.ResetInstance()
			},
			servicesEnv: "NonExistentService",
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			if tt.servicesEnv != "" {
				t.Setenv(envVarNameServicesEnabled, tt.servicesEnv)
			}

			if tt.expectPanic {
				assert.Panics(t, func() {
					Init()
				})
				return
			}

			// Should not panic
			assert.NotPanics(t, func() {
				Init()
			})

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}