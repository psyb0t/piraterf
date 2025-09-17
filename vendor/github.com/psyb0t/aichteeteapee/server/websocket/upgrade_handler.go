package websocket

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

// UpgradeHandler creates an HTTP handler that upgrades connections to WebSocket
func UpgradeHandler( //nolint:funlen
	hub Hub,
	opts ...HandlerOption,
) http.HandlerFunc {
	config := NewHandlerConfig() // Start with defaults

	// Apply user options to override defaults
	for _, opt := range opts {
		opt(&config)
	}

	// Create WebSocket upgrader with config
	upgrader := websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		HandshakeTimeout:  config.HandshakeTimeout,
		CheckOrigin:       config.CheckOrigin,
		Subprotocols:      config.Subprotocols,
		EnableCompression: config.EnableCompression,
	}

	logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldReadBufferSize:    config.ReadBufferSize,
		aichteeteapee.FieldWriteBufferSize:   config.WriteBufferSize,
		aichteeteapee.FieldHandshakeTimeout:  config.HandshakeTimeout,
		aichteeteapee.FieldEnableCompression: config.EnableCompression,
		aichteeteapee.FieldHubName:           hub.Name(),
	}).Debug("created websocket upgrade handler")

	return func(w http.ResponseWriter, r *http.Request) {
		logger := logrus.WithFields(logrus.Fields{
			aichteeteapee.FieldRemoteAddr: r.RemoteAddr,
			aichteeteapee.FieldOrigin:     r.Header.Get("Origin"),
		})

		logger.WithFields(logrus.Fields{
			aichteeteapee.FieldUserAgent: r.UserAgent(),
			aichteeteapee.FieldEndpoint:  r.URL.Path,
		}).Debug("websocket upgrade request received")

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.WithError(err).Error("websocket upgrade failed")

			return // upgrader already wrote HTTP error response
		}

		// Handle connection close before any other operations
		conn.SetCloseHandler(func(code int, text string) error {
			logger.WithFields(logrus.Fields{
				aichteeteapee.FieldCloseCode: code,
				aichteeteapee.FieldCloseText: text,
			}).Debug("websocket close handler triggered")

			return nil
		})

		logger.Info("websocket connection established")

		// Extract client ID and get or create client
		var client *Client

		clientID := extractClientIDFromRequest(r)
		if clientID != "" {
			parsedClientID := parseClientID(clientID)
			clientLogger := logger.WithField(aichteeteapee.FieldClientID, parsedClientID)

			// Use atomic get-or-create to avoid race conditions
			var wasCreated bool

			client, wasCreated = hub.GetOrCreateClient(parsedClientID, config.ClientOptions...)

			if !wasCreated {
				clientLogger.Debug("adding connection to existing client")
			} else {
				clientLogger.Debug("created new client with specified ID")
			}

			// Create and add connection using AddConnection
			connection := NewConnection(conn, client)
			client.AddConnection(connection)
		} else {
			logger.Debug("creating client with generated ID")

			client = NewClient(config.ClientOptions...)
			hub.AddClient(client)

			// Create and add connection using AddConnection
			connection := NewConnection(conn, client)
			client.AddConnection(connection)
		}

		finalLogger := logger.WithField(aichteeteapee.FieldClientID, client.ID())
		finalLogger.Debug("client ready")
		finalLogger.Debug("client connection handled successfully")
	}
}

// extractClientIDFromRequest extracts client ID from request
// This can be customized based on your authentication system
func extractClientIDFromRequest(r *http.Request) string {
	// Try to extract from query parameter first
	if clientID := r.URL.Query().Get("clientID"); clientID != "" {
		return clientID
	}

	// Try to extract from custom header
	if clientID := r.Header.Get(aichteeteapee.HeaderNameXClientID); clientID != "" {
		return clientID
	}

	// Could also extract from JWT token, session, cookies, etc.
	// For now, return empty string to generate a new client ID
	return ""
}

// parseClientID converts string client ID to UUID
// Returns zero UUID if parsing fails, which will generate a new UUID
func parseClientID(clientID string) uuid.UUID {
	if parsedID, err := uuid.Parse(clientID); err == nil {
		return parsedID
	}

	logrus.WithField("providedClientID", clientID).Warn("invalid client ID format, generating new UUID")

	return uuid.New()
}
