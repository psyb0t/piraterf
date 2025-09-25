package wsunixbridge

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/psyb0t/aichteeteapee"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/sirupsen/logrus"
)

const (
	dirPermissions = 0o750
	bufferSize     = 4096

	// Socket constants.
	writerUnixSockSuffix = "_output"
	readerUnixSockSuffix = "_input"

	// Event type for initialization.
	EventTypeWSUnixBridgeInitialized dabluveees.EventType = "wsunixbridge.init"
)

// NewUpgradeHandler creates a new WebSocket Unix socket upgrade handler.
func NewUpgradeHandler(
	socketsDir string,
	connHandler ConnectionHandler,
) http.HandlerFunc {
	config := dabluveees.NewUpgradeHandlerConfig()

	upgrader := websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		HandshakeTimeout:  config.HandshakeTimeout,
		CheckOrigin:       config.CheckOrigin,
		Subprotocols:      config.Subprotocols,
		EnableCompression: config.EnableCompression,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		handleConnection(w, r, socketsDir, connHandler, upgrader)
	}
}

func handleConnection(
	w http.ResponseWriter,
	r *http.Request,
	socketsDir string,
	connHandler ConnectionHandler,
	upgrader websocket.Upgrader,
) {
	connID := uuid.New()
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldRemoteAddr:   r.RemoteAddr,
		aichteeteapee.FieldOrigin:       r.Header.Get(aichteeteapee.HeaderNameOrigin),
		aichteeteapee.FieldConnectionID: connID,
	})

	logger.Debug("unixsock websocket upgrade request received")

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("websocket upgrade failed")

		return // upgrader already wrote HTTP error response
	}

	logger.Info("unixsock websocket connection established")

	err = setupConnection(
		r.Context(),
		wsConn,
		socketsDir,
		connID,
		connHandler,
		logger,
	)
	if err != nil {
		logger.WithError(err).Error("connection setup failed")

		if err := wsConn.Close(); err != nil {
			logger.WithError(err).Debug("error closing websocket connection")
		}
	}
}

func handleWebSocketMessages(
	wsConn *websocket.Conn,
	conn *Connection,
	logger *logrus.Entry,
) {
	logger.Debugf(
		"handling websocket messages for connection %s",
		conn.ID,
	)

	defer logger.Debugf(
		"finished handling websocket messages for connection %s",
		conn.ID,
	)

	for {
		messageType, data, err := wsConn.ReadMessage()
		if err != nil {
			isCloseError := websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
			)

			if isCloseError {
				logger.Info("websocket connection closed normally")

				return
			}

			logger.WithError(err).Error("websocket read error")

			return
		}

		if messageType != websocket.BinaryMessage &&
			messageType != websocket.TextMessage {
			continue
		}

		// Broadcast to all connected output readers
		conn.WriterUnixSock.Broadcast(data, logger)
	}
}
