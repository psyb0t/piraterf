package wsunixbridge

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/psyb0t/aichteeteapee"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

// UnixSock represents Unix socket resources for a connection.
type UnixSock struct {
	Listener   net.Listener
	Clients    []net.Conn
	ClientsMux sync.RWMutex
	Path       string // Full path to the Unix socket file
}

// Broadcast sends data to all connected clients.
func (us *UnixSock) Broadcast(data []byte, logger *logrus.Entry) {
	logger.Debug("collecting clients to broadcast")

	us.ClientsMux.RLock()
	clients := make([]net.Conn, len(us.Clients))
	copy(clients, us.Clients)
	us.ClientsMux.RUnlock()

	logger.Debugf("broadcasting %d clients", len(clients))

	for _, client := range clients {
		if _, err := client.Write(data); err != nil {
			logger.WithError(err).
				Debug("failed to write to UnixSock client")
		}
	}

	logger.Debug("broadcast complete")
}

func createUnixSockets(
	ctx context.Context,
	socketsDir string,
	connID uuid.UUID,
	conn *Connection,
	logger *logrus.Entry,
) error {
	basePath := filepath.Join(socketsDir, connID.String())
	outputPath := basePath + writerUnixSockSuffix
	inputPath := basePath + readerUnixSockSuffix

	// Remove any existing sockets
	if err := os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
		logger.WithError(err).WithField(aichteeteapee.FieldPath, outputPath).
			Debug("error removing existing WriterUnixSock socket")
	}

	if err := os.Remove(inputPath); err != nil && !os.IsNotExist(err) {
		logger.WithError(err).WithField(aichteeteapee.FieldPath, inputPath).
			Debug("error removing existing ReaderUnixSock socket")
	}

	lc := &net.ListenConfig{}

	// Create output socket (external tools read WebSocket data from here)
	writerUnixSockListener, err := lc.Listen(
		ctx,
		aichteeteapee.NetworkTypeUnix,
		outputPath,
	)
	if err != nil {
		return ctxerrors.Wrap(err, "failed to create WriterUnixSock socket")
	}

	conn.WriterUnixSock.Listener = writerUnixSockListener
	conn.WriterUnixSock.Path = outputPath

	// Create input socket (external tools write data here to send to WebSocket)
	readerUnixSockListener, err := lc.Listen(
		ctx,
		aichteeteapee.NetworkTypeUnix,
		inputPath,
	)
	if err != nil {
		return ctxerrors.Wrap(err, "failed to create ReaderUnixSock socket")
	}

	conn.ReaderUnixSock.Listener = readerUnixSockListener
	conn.ReaderUnixSock.Path = inputPath

	return nil
}

func acceptWriterUnixSockClients(
	ctx context.Context,
	conn *Connection,
	logger *logrus.Entry,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			client, err := conn.WriterUnixSock.Listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				logger.WithError(err).Debug("WriterUnixSock socket accept error")

				continue
			}

			conn.WriterUnixSock.ClientsMux.Lock()
			conn.WriterUnixSock.Clients = append(conn.WriterUnixSock.Clients, client)
			conn.WriterUnixSock.ClientsMux.Unlock()

			logger.Debug("new WriterUnixSock client connected")

			// Handle client disconnection
			go handleWriterUnixSockClient(ctx, conn, client, logger)
		}
	}
}

func handleWriterUnixSockClient(
	ctx context.Context,
	conn *Connection,
	client net.Conn,
	logger *logrus.Entry,
) {
	defer func() {
		if err := client.Close(); err != nil {
			logger.WithError(err).Debug("error closing WriterUnixSock client connection")
		}

		removeWriterUnixSockClient(conn, client)
		logger.Debug("WriterUnixSock client disconnected")
	}()

	buffer := make([]byte, 1)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := client.Read(buffer)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				logger.WithError(err).Debug("WriterUnixSock client read error")

				return
			}
		}
	}
}

func acceptReaderUnixSockClients(
	ctx context.Context,
	conn *Connection,
	logger *logrus.Entry,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			client, err := conn.ReaderUnixSock.Listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				logger.WithError(err).Debug("ReaderUnixSock socket accept error")

				continue
			}

			conn.ReaderUnixSock.ClientsMux.Lock()
			conn.ReaderUnixSock.Clients = append(conn.ReaderUnixSock.Clients, client)
			conn.ReaderUnixSock.ClientsMux.Unlock()

			logger.Debug("new ReaderUnixSock client connected")

			go handleReaderUnixSockClient(ctx, conn, client, logger)
		}
	}
}

func handleReaderUnixSockClient(
	ctx context.Context,
	conn *Connection,
	client net.Conn,
	logger *logrus.Entry,
) {
	defer func() {
		if err := client.Close(); err != nil {
			logger.WithError(err).Debug("error closing ReaderUnixSock client connection")
		}

		removeReaderUnixSockClient(conn, client)
		logger.Debug("ReaderUnixSock client disconnected")
	}()

	buffer := make([]byte, bufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := client.Read(buffer)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				logger.WithError(err).Debug("error reading from ReaderUnixSock socket")

				return
			}

			if n == 0 {
				continue
			}

			err = conn.Conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
			if err != nil {
				logger.WithError(err).Debug("error writing to websocket")

				return
			}
		}
	}
}

func removeReaderUnixSockClient(conn *Connection, client net.Conn) {
	conn.ReaderUnixSock.ClientsMux.Lock()
	defer conn.ReaderUnixSock.ClientsMux.Unlock()

	for i, connClient := range conn.ReaderUnixSock.Clients {
		if connClient == client {
			conn.ReaderUnixSock.Clients = append(
				conn.ReaderUnixSock.Clients[:i],
				conn.ReaderUnixSock.Clients[i+1:]...,
			)

			break
		}
	}
}

func removeWriterUnixSockClient(conn *Connection, client net.Conn) {
	conn.WriterUnixSock.ClientsMux.Lock()
	defer conn.WriterUnixSock.ClientsMux.Unlock()

	for i, connClient := range conn.WriterUnixSock.Clients {
		if connClient == client {
			conn.WriterUnixSock.Clients = append(
				conn.WriterUnixSock.Clients[:i],
				conn.WriterUnixSock.Clients[i+1:]...,
			)

			break
		}
	}
}

func removeSocketFiles(conn *Connection, logger *logrus.Entry) {
	// Remove output socket file
	if conn.WriterUnixSock.Listener != nil {
		if ul, ok := conn.WriterUnixSock.Listener.(*net.UnixListener); ok {
			if err := os.Remove(ul.Addr().String()); err != nil && !os.IsNotExist(err) {
				logger.WithError(err).
					WithField(aichteeteapee.FieldPath, ul.Addr().String()).
					Debug("error removing WriterUnixSock socket file")
			}
		}
	}

	// Remove input socket file
	if conn.ReaderUnixSock.Listener != nil {
		if ul, ok := conn.ReaderUnixSock.Listener.(*net.UnixListener); ok {
			if err := os.Remove(ul.Addr().String()); err != nil && !os.IsNotExist(err) {
				logger.WithError(err).
					WithField(aichteeteapee.FieldPath, ul.Addr().String()).
					Debug("error removing ReaderUnixSock socket file")
			}
		}
	}
}
