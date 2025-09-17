package websocket

import (
	"net/http"
	"time"

	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

// HandlerConfig holds WebSocket handler configuration
type HandlerConfig struct {
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	CheckOrigin       func(*http.Request) bool
	Subprotocols      []string
	EnableCompression bool
	ClientOptions     []ClientOption
}

// NewHandlerConfig creates config with defaults from http/defaults.go
func NewHandlerConfig() HandlerConfig {
	config := HandlerConfig{
		ReadBufferSize:    aichteeteapee.DefaultWebSocketHandlerReadBufferSize,
		WriteBufferSize:   aichteeteapee.DefaultWebSocketHandlerWriteBufferSize,
		HandshakeTimeout:  aichteeteapee.DefaultWebSocketHandlerHandshakeTimeout,
		EnableCompression: aichteeteapee.DefaultWebSocketHandlerEnableCompression,
		Subprotocols:      []string{},
		ClientOptions:     []ClientOption{},
		CheckOrigin:       aichteeteapee.GetDefaultWebSocketCheckOrigin,
	}

	logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldReadBufferSize:    config.ReadBufferSize,
		aichteeteapee.FieldWriteBufferSize:   config.WriteBufferSize,
		aichteeteapee.FieldHandshakeTimeout:  config.HandshakeTimeout,
		aichteeteapee.FieldEnableCompression: config.EnableCompression,
	}).Debug("created websocket handler config with defaults")

	return config
}

type HandlerOption func(*HandlerConfig)

// WithHandlerBufferSizes sets both read and write buffer sizes for the WebSocket upgrader
func WithHandlerBufferSizes(read, write int) HandlerOption {
	return func(c *HandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldReadSize":  c.ReadBufferSize,
			"oldWriteSize": c.WriteBufferSize,
			"newReadSize":  read,
			"newWriteSize": write,
		}).Debug("updating handler buffer sizes")

		c.ReadBufferSize = read
		c.WriteBufferSize = write
	}
}

// WithHandshakeTimeout sets the WebSocket handshake timeout
func WithHandshakeTimeout(timeout time.Duration) HandlerOption {
	return func(c *HandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldTimeout": c.HandshakeTimeout,
			"newTimeout": timeout,
		}).Debug("updating handler handshake timeout")

		c.HandshakeTimeout = timeout
	}
}

// WithCompression enables or disables WebSocket compression
func WithCompression(enable bool) HandlerOption {
	return func(c *HandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldCompression": c.EnableCompression,
			"newCompression": enable,
		}).Debug("updating handler compression setting")

		c.EnableCompression = enable
	}
}

// WithSubprotocols sets the supported WebSocket subprotocols
func WithSubprotocols(protocols ...string) HandlerOption {
	return func(c *HandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldProtocols": c.Subprotocols,
			"newProtocols": protocols,
		}).Debug("updating handler subprotocols")

		c.Subprotocols = protocols
	}
}

// WithCheckOrigin sets the origin checking function for WebSocket connections
func WithCheckOrigin(checkOrigin func(*http.Request) bool) HandlerOption {
	return func(c *HandlerConfig) {
		logrus.Debug("updating handler CheckOrigin function")

		c.CheckOrigin = checkOrigin
	}
}

// WithHandlerClientOptions adds default client options that will be applied to all new clients
func WithHandlerClientOptions(opts ...ClientOption) HandlerOption {
	return func(c *HandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"existingOptionsCount": len(c.ClientOptions),
			"newOptionsCount":      len(opts),
		}).Debug("adding handler client options")

		c.ClientOptions = append(c.ClientOptions, opts...)
	}
}
