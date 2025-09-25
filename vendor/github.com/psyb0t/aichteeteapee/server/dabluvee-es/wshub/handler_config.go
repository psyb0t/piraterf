package wshub

import (
	"net/http"
	"time"

	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
)

// UpgradeHandlerConfig extends the shared websocket config
// with hub-specific options.
type UpgradeHandlerConfig struct {
	dabluveees.UpgradeHandlerConfig
	ClientOptions []ClientOption
}

// NewUpgradeHandlerConfig creates config with defaults.
func NewUpgradeHandlerConfig() UpgradeHandlerConfig {
	return UpgradeHandlerConfig{
		UpgradeHandlerConfig: dabluveees.NewUpgradeHandlerConfig(),
		ClientOptions:        []ClientOption{},
	}
}

type UpgradeHandlerOption func(*UpgradeHandlerConfig)

// WithUpgradeHandlerClientOptions adds default client options
// that will be applied to all new clients.
func WithUpgradeHandlerClientOptions(
	opts ...ClientOption,
) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		c.ClientOptions = append(c.ClientOptions, opts...)
	}
}

// Wrapper functions for shared websocket options.
func WithUpgradeHandlerBufferSizes(read, write int) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		dabluveees.WithUpgradeHandlerBufferSizes(read, write)(&c.UpgradeHandlerConfig)
	}
}

func WithUpgradeHandlerHandshakeTimeout(
	timeout time.Duration,
) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		dabluveees.WithUpgradeHandlerHandshakeTimeout(timeout)(
			&c.UpgradeHandlerConfig,
		)
	}
}

func WithUpgradeHandlerCompression(enable bool) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		dabluveees.WithUpgradeHandlerCompression(enable)(&c.UpgradeHandlerConfig)
	}
}

func WithUpgradeHandlerSubprotocols(protocols ...string) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		dabluveees.WithUpgradeHandlerSubprotocols(protocols...)(
			&c.UpgradeHandlerConfig,
		)
	}
}

func WithUpgradeHandlerCheckOrigin(
	checkOrigin func(*http.Request) bool,
) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		dabluveees.WithUpgradeHandlerCheckOrigin(checkOrigin)(&c.UpgradeHandlerConfig)
	}
}
