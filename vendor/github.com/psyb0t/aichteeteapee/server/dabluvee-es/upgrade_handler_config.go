package dabluveees

import (
	"net/http"
	"time"

	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

// UpgradeHandlerConfig holds WebSocket upgrade handler configuration.
type UpgradeHandlerConfig struct {
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	CheckOrigin       func(*http.Request) bool
	Subprotocols      []string
	EnableCompression bool
}

// NewUpgradeHandlerConfig creates config with defaults from http/defaults.go.
func NewUpgradeHandlerConfig() UpgradeHandlerConfig {
	config := UpgradeHandlerConfig{
		ReadBufferSize:    aichteeteapee.DefaultWebSocketHandlerReadBufferSize,
		WriteBufferSize:   aichteeteapee.DefaultWebSocketHandlerWriteBufferSize,
		HandshakeTimeout:  aichteeteapee.DefaultWebSocketHandlerHandshakeTimeout,
		EnableCompression: aichteeteapee.DefaultWebSocketHandlerEnableCompression,
		Subprotocols:      []string{},
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
