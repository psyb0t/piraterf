package wsunixbridge

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	commontypes "github.com/psyb0t/common-go/types"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

// Connection represents a WebSocket connection with Unix socket streams.
type Connection struct {
	ID             uuid.UUID
	Conn           *websocket.Conn
	WriterUnixSock UnixSock // Unix socket for writing data
	ReaderUnixSock UnixSock // Unix socket for reading data
}

// ConnectionHandler is called when a new connection is established.
type ConnectionHandler func(connection *Connection) error

// InitMessageData represents the data sent in the initialization event.
type InitMessageData struct {
	WriterSocket string `json:"writerSocket"`
	ReaderSocket string `json:"readerSocket"`
}

// Global map to track active connections.
//
//nolint:gochecknoglobals
var connectionSockets = commontypes.NewMapWithMutex[uuid.UUID, *Connection]()

func setupConnection( //nolint:funlen
	ctx context.Context,
	wsConn *websocket.Conn,
	socketsDir string,
	connID uuid.UUID,
	connHandler ConnectionHandler,
	logger *logrus.Entry,
) error {
	if err := os.MkdirAll(socketsDir, dirPermissions); err != nil {
		return ctxerrors.Wrap(err, "failed to create sockets directory")
	}

	conn := &Connection{
		ID:   connID,
		Conn: wsConn,
	}

	if err := createUnixSockets(
		ctx, socketsDir, connID, conn, logger,
	); err != nil {
		return err
	}

	logger.Info("created connection Unix sockets",
		"outputPath", conn.WriterUnixSock.Path,
		"inputPath", conn.ReaderUnixSock.Path,
	)

	// Send initialization event to client with socket paths
	initData := InitMessageData{
		WriterSocket: conn.WriterUnixSock.Path,
		ReaderSocket: conn.ReaderUnixSock.Path,
	}
	initEvent := dabluveees.NewEvent(EventTypeWSUnixBridgeInitialized, initData)

	if err := wsConn.WriteJSON(initEvent); err != nil {
		logger.WithError(err).Error("failed to send initialization event")

		return ctxerrors.Wrap(err, "failed to send initialization event")
	}

	logger.Info("sent wsunixbridge initialization event to client")

	// Start socket servers - use background context to prevent cancellation
	// when HTTP request ends
	serverCtx, cancel := context.WithCancel(context.Background())

	//nolint:contextcheck // Intentional background context for accept goroutines
	go acceptWriterUnixSockClients(serverCtx, conn, logger)
	//nolint:contextcheck // Intentional background context for accept goroutines
	go acceptReaderUnixSockClients(serverCtx, conn, logger)

	// Call user handler
	if connHandler != nil {
		go func() {
			if err := connHandler(conn); err != nil {
				logger.WithError(err).Error("connection handler error")
			}
		}()
	}

	// Handle WebSocket messages and connection lifecycle in goroutine
	go func() {
		// Store connection
		connectionSockets.Set(connID, conn)

		// Handle connection cleanup when goroutine exits
		defer func() {
			cancel() // Cancel socket servers
			removeConnection(connID, logger)
		}()

		// Handle WebSocket messages (blocks until connection closes)
		handleWebSocketMessages(wsConn, conn, logger)
	}()

	return nil
}

func removeConnection(connID uuid.UUID, logger *logrus.Entry) {
	conn, exists := connectionSockets.Get(connID)
	if !exists {
		return
	}

	closeAllClients(conn, logger)
	closeListeners(conn, logger)
	removeSocketFiles(conn, logger)

	connectionSockets.Delete(connID)
	logger.Info("removed connection Unix sockets")
}

func closeAllClients(conn *Connection, logger *logrus.Entry) {
	// Close output readers
	conn.WriterUnixSock.ClientsMux.Lock()

	for _, client := range conn.WriterUnixSock.Clients {
		if err := client.Close(); err != nil {
			logger.WithError(err).
				Debug("error closing WriterUnixSock client connection during cleanup")
		}
	}

	conn.WriterUnixSock.ClientsMux.Unlock()

	// Close input writers
	conn.ReaderUnixSock.ClientsMux.Lock()

	for _, client := range conn.ReaderUnixSock.Clients {
		if err := client.Close(); err != nil {
			logger.WithError(err).
				Debug("error closing ReaderUnixSock client connection during cleanup")
		}
	}

	conn.ReaderUnixSock.ClientsMux.Unlock()
}

func closeListeners(conn *Connection, logger *logrus.Entry) {
	// Close output listener
	if conn.WriterUnixSock.Listener != nil {
		if err := conn.WriterUnixSock.Listener.Close(); err != nil {
			logger.WithError(err).Debug("error closing WriterUnixSock listener")
		}
	}

	// Close input listener
	if conn.ReaderUnixSock.Listener != nil {
		if err := conn.ReaderUnixSock.Listener.Close(); err != nil {
			logger.WithError(err).Debug("error closing ReaderUnixSock listener")
		}
	}
}
