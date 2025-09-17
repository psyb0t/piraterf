package websocket

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

type Connection struct {
	id       uuid.UUID       // UUID4 connection identifier
	conn     *websocket.Conn // WebSocket connection
	client   *Client         // Reference to parent client
	sendCh   chan *Event     // Per-connection message channel
	doneCh   chan struct{}   // Connection shutdown signal
	stopOnce sync.Once       // Ensure single stop
	isDone   atomic.Bool     // Atomic flag for connection state
	sendWg   sync.WaitGroup  // Wait for in-flight sends to complete
}

// NewConnection creates a new WebSocket connection
func NewConnection(
	conn *websocket.Conn,
	client *Client,
) *Connection {
	return &Connection{
		id:       uuid.New(),
		conn:     conn,
		client:   client,
		sendCh:   make(chan *Event, client.config.SendBufferSize),
		doneCh:   make(chan struct{}),
		stopOnce: sync.Once{},
	}
}

// GetHubName safely returns the hub name, handling nil cases
func (c *Connection) GetHubName() string {
	if c.client == nil {
		return "unknown"
	}

	return c.client.GetHubName()
}

// GetClientID safely returns the client ID, handling nil cases
func (c *Connection) GetClientID() uuid.UUID {
	if c.client == nil {
		return uuid.Nil
	}

	return c.client.id
}

// Send sends an event to the connection's send channel
func (c *Connection) Send(event *Event) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:      c.GetHubName(),
		aichteeteapee.FieldClientID:     c.GetClientID(),
		aichteeteapee.FieldConnectionID: c.id,
		aichteeteapee.FieldEventType:    event.Type,
		aichteeteapee.FieldEventID:      event.ID,
	})

	if c.isDone.Load() {
		logger.Debug("connection is done, cannot send event")

		return
	}

	if c.sendCh == nil || c.doneCh == nil {
		logger.Debug("connection channels are nil, cannot send event")

		return
	}

	// Add to WaitGroup to track this send operation
	c.sendWg.Add(1)
	defer c.sendWg.Done()

	// Check again after acquiring WaitGroup (race protection)
	if c.isDone.Load() {
		logger.Debug("connection became done during send, aborting")

		return
	}

	select {
	case c.sendCh <- event:
		logger.Debug("event queued for sending")
	case <-c.doneCh:
		logrus.WithFields(logrus.Fields{
			aichteeteapee.FieldHubName:      c.GetHubName(),
			aichteeteapee.FieldClientID:     c.GetClientID(),
			aichteeteapee.FieldConnectionID: c.id,
		}).Debug("connection stopped, cannot send event")
	}
}

// Stop cleanly shuts down the connection
func (c *Connection) Stop() {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:      c.GetHubName(),
		aichteeteapee.FieldClientID:     c.GetClientID(),
		aichteeteapee.FieldConnectionID: c.id,
	})

	c.stopOnce.Do(func() {
		// Set done flag atomically first
		c.isDone.Store(true)

		logger.Debug("stopping connection")

		// Close doneCh first to signal shutdown
		if c.doneCh != nil {
			close(c.doneCh)
		}

		// Wait for all in-flight sends to complete before closing sendCh
		logger.Debug("waiting for in-flight sends to complete")
		c.sendWg.Wait()

		// Now safe to close sendCh - no more sends can happen
		if c.sendCh != nil {
			close(c.sendCh)
		}

		if c.conn != nil {
			_ = c.conn.Close()
		}

		logger.Debug("connection stop completed")
	})
}

// writePump handles outbound messages and keepalive
func (c *Connection) writePump() { //nolint:cyclop,funlen
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:      c.GetHubName(),
		aichteeteapee.FieldClientID:     c.GetClientID(),
		aichteeteapee.FieldConnectionID: c.id,
	})

	defer func() {
		logger.Debug("write pump stopped")
		c.client.RemoveConnection(c.id) // Auto-cleanup
	}()

	if c.isDone.Load() {
		logger.Debug("connection is done, cannot start write pump")

		return
	}

	logger.Debug("starting connection write pump")

	ticker := time.NewTicker(c.client.config.PingInterval)
	defer ticker.Stop()

	for {
		if c.isDone.Load() {
			logger.Debug("connection is done, stopping write pump")

			return
		}

		select {
		case event, ok := <-c.sendCh:
			if !ok {
				// Channel closed, send close message
				_ = c.conn.SetWriteDeadline(time.Now().Add(c.client.config.WriteTimeout))
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

				return
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(c.client.config.WriteTimeout))

			if err := c.conn.WriteJSON(event); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Info("connection closed normally during write")

					return
				}

				logger.WithError(err).Error("connection write error")

				return
			}

			logger.WithFields(logrus.Fields{
				aichteeteapee.FieldEventType: string(event.Type),
				aichteeteapee.FieldEventID:   event.ID,
			}).Debug("event sent to connection")

		case <-ticker.C:
			// Send ping
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.client.config.WriteTimeout))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Info("connection closed normally during ping")

					return
				}

				logger.WithError(err).Error("connection ping error")

				return
			}

		case <-c.doneCh:
			return
		}
	}
}

// readPump handles inbound messages and connection monitoring
func (c *Connection) readPump() { //nolint:cyclop,funlen
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:      c.GetHubName(),
		aichteeteapee.FieldClientID:     c.GetClientID(),
		aichteeteapee.FieldConnectionID: c.id,
	})

	defer func() {
		logger.Debug("read pump stopped")
		c.client.RemoveConnection(c.id) // Auto-cleanup
	}()

	if c.isDone.Load() {
		logger.Debug("connection is done, cannot start read pump")

		return
	}

	logger.Debug("starting connection read pump")

	// Configure connection
	c.conn.SetReadLimit(c.client.config.ReadLimit)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.client.config.ReadTimeout))

	// Set pong handler
	c.conn.SetPongHandler(func(string) error {
		if c.isDone.Load() {
			return websocket.ErrCloseSent
		}

		_ = c.conn.SetReadDeadline(time.Now().Add(c.client.config.PongTimeout))

		return nil
	})

	for {
		if c.isDone.Load() {
			logger.Debug("connection is done, stopping read pump")

			return
		}

		select {
		case <-c.doneCh:
			return
		default:
		}

		var event Event
		if err := c.conn.ReadJSON(&event); err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				logger.Info("websocket connection closed")

				return
			}

			logger.WithError(err).Error("connection read error")

			return
		}

		// Check if done after reading event
		if c.isDone.Load() {
			logger.WithFields(logrus.Fields{
				aichteeteapee.FieldEventType: string(event.Type),
				aichteeteapee.FieldEventID:   event.ID,
			}).Debug("connection is done, cannot process event")

			return
		}

		// Reset read deadline
		_ = c.conn.SetReadDeadline(time.Now().Add(c.client.config.ReadTimeout))

		logger.WithFields(logrus.Fields{
			aichteeteapee.FieldEventType: string(event.Type),
			aichteeteapee.FieldEventID:   event.ID,
		}).Debug("event received from connection")

		// Process event through hub
		if c.client.hub != nil {
			c.client.hub.ProcessEvent(c.client, &event)

			continue
		}

		logger.Debug("hub is nil, cannot process event")
	}
}
