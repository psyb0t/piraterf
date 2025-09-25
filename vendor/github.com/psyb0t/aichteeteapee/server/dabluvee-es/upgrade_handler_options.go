package dabluveees

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type UpgradeHandlerOption func(*UpgradeHandlerConfig)

// WithUpgradeHandlerBufferSizes sets both read and write buffer sizes
// for the WebSocket upgrader.
func WithUpgradeHandlerBufferSizes(read, write int) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
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

// WithUpgradeHandlerHandshakeTimeout sets the WebSocket handshake timeout.
func WithUpgradeHandlerHandshakeTimeout(
	timeout time.Duration,
) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldTimeout": c.HandshakeTimeout,
			"newTimeout": timeout,
		}).Debug("updating handler handshake timeout")

		c.HandshakeTimeout = timeout
	}
}

// WithUpgradeHandlerCompression enables or disables WebSocket compression.
func WithUpgradeHandlerCompression(enable bool) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldCompression": c.EnableCompression,
			"newCompression": enable,
		}).Debug("updating handler compression setting")

		c.EnableCompression = enable
	}
}

// WithUpgradeHandlerSubprotocols sets the supported WebSocket subprotocols.
func WithUpgradeHandlerSubprotocols(protocols ...string) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		logrus.WithFields(logrus.Fields{
			"oldProtocols": c.Subprotocols,
			"newProtocols": protocols,
		}).Debug("updating handler subprotocols")

		c.Subprotocols = protocols
	}
}

// WithUpgradeHandlerCheckOrigin sets the origin checking function for
// WebSocket connections.
func WithUpgradeHandlerCheckOrigin(
	checkOrigin func(*http.Request) bool,
) UpgradeHandlerOption {
	return func(c *UpgradeHandlerConfig) {
		logrus.Debug("updating handler CheckOrigin function")

		c.CheckOrigin = checkOrigin
	}
}
