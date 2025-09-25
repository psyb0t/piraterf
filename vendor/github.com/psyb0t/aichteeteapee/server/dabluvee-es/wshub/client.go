package wshub

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/sirupsen/logrus"
)

type Client struct {
	id            uuid.UUID              // UUID4 client identifier
	hub           Hub                    // Reference to hub
	hubMu         sync.RWMutex           // Protects hub field
	connections   *connectionsMap        // Thread-safe connection management
	sendCh        chan *dabluveees.Event // Client-level message channel
	doneCh        chan struct{}          // Client shutdown signal
	wg            sync.WaitGroup         // Wait for goroutines to finish
	stopOnce      sync.Once              // Ensure single stop
	config        ClientConfig           // Client configuration
	isStopped     atomic.Bool            // Atomic flag for client stopped state
	isRunning     atomic.Bool            // Atomic flag for client running state
	readyToStopCh chan struct{}          // Channel to signal ready to stop
}

// GetHubName safely returns the hub name, handling nil cases.
func (c *Client) GetHubName() string {
	c.hubMu.RLock()
	defer c.hubMu.RUnlock()

	if c.hub == nil {
		return "unknown"
	}

	return c.hub.Name()
}

func NewClient(opts ...ClientOption) *Client {
	return NewClientWithID(uuid.New(), opts...)
}

func NewClientWithID(clientID uuid.UUID, opts ...ClientOption) *Client {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldClientID: clientID,
	})

	config := NewClientConfig()
	for _, opt := range opts {
		opt(&config)
	}

	client := &Client{
		id:            clientID,
		hub:           nil, // Hub will be set when added to hub
		connections:   newConnectionsMap(),
		sendCh:        make(chan *dabluveees.Event, config.SendBufferSize),
		doneCh:        make(chan struct{}),
		readyToStopCh: make(chan struct{}),
		config:        config,
	}

	logger.Debug("created new client")

	return client
}

func (c *Client) ID() uuid.UUID {
	return c.id
}

func (c *Client) SetHub(hub Hub) {
	c.hubMu.Lock()
	defer c.hubMu.Unlock()

	c.hub = hub
}

func (c *Client) AddConnection(conn *Connection) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:      c.GetHubName(),
		aichteeteapee.FieldClientID:     c.id,
		aichteeteapee.FieldConnectionID: conn.id,
	})

	if c.isStopped.Load() {
		logger.Debug("client is done, cannot add connection")

		return
	}

	logger.WithField(aichteeteapee.FieldTotalConns, c.connections.Count()+1).
		Debug("adding new connection to client")

	c.connections.Add(conn)

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		if conn.conn != nil {
			conn.readPump()
		}
	}()

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		if conn.conn != nil {
			conn.writePump()
		}
	}()

	logger.Debug("connection pumps started")
}

func (c *Client) RemoveConnection(connectionID uuid.UUID) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:      c.GetHubName(),
		aichteeteapee.FieldClientID:     c.id,
		aichteeteapee.FieldConnectionID: connectionID,
	})

	if c.isStopped.Load() {
		logger.Debug("client is done, cannot remove connection")

		return
	}

	conn := c.connections.Remove(connectionID)
	if conn != nil {
		connectionCount := c.connections.Count()
		logger.WithField(aichteeteapee.FieldTotalConns, connectionCount).
			Debug("removed connection from client")

		conn.Stop()

		// If no connections left, trigger client shutdown
		if connectionCount == 0 {
			logger.Debug("no connections left, triggering client shutdown")

			go c.Stop()
		}
	}
}

func (c *Client) GetConnections() map[uuid.UUID]*Connection {
	if c.connections == nil {
		return make(map[uuid.UUID]*Connection)
	}

	return c.connections.GetAll()
}

func (c *Client) ConnectionCount() int {
	if c.connections == nil {
		return 0
	}

	return c.connections.Count()
}

// SendEvent sends an event to all client connections
// (alias for Send for hub compatibility).
func (c *Client) SendEvent(event *dabluveees.Event) {
	c.Send(event)
}

// Send sends an event to the client's send channel for distribution.
func (c *Client) Send(event *dabluveees.Event) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:   c.GetHubName(),
		aichteeteapee.FieldClientID:  c.id,
		aichteeteapee.FieldEventType: event.Type,
		aichteeteapee.FieldEventID:   event.ID,
	})

	if c.isStopped.Load() {
		logger.Debug("client is done, cannot send event")

		return
	}

	if c.sendCh == nil || c.doneCh == nil {
		logger.Debug("client channels are nil, cannot send event")

		return
	}

	select {
	case c.sendCh <- event:
		logger.Debug("event queued for client distribution")
	case <-c.doneCh:
		logger.Debug("client stopped, cannot send event")
	default:
		logger.WithField(aichteeteapee.FieldBufferSize, cap(c.sendCh)).
			Warn("client send buffer full, dropping message")
	}
}

// IsSubscribedTo checks if client is subscribed to an event type.
func (c *Client) IsSubscribedTo(_ dabluveees.EventType) bool {
	return true // For now, accept all events
}

// Stop gracefully shuts down the client and all its connections.
func (c *Client) Stop() {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:  c.GetHubName(),
		aichteeteapee.FieldClientID: c.id,
	})

	// Bail out if Run() was never called
	if !c.isRunning.Load() {
		return
	}

	// Wait for Run() to be ready for stopping
	<-c.readyToStopCh

	c.stopOnce.Do(func() {
		// Set done flag atomically first
		c.isStopped.Store(true)

		logger.Info("stopping client")

		// Signal shutdown - check for nil channels first
		if c.doneCh != nil {
			close(c.doneCh)
		}

		if c.sendCh != nil {
			close(c.sendCh)
		}

		// Stop all connections - check for nil connections map first
		if c.connections != nil {
			for _, conn := range c.connections.GetAll() {
				conn.Stop()
			}
		}

		// Wait for goroutines to finish with timeout
		done := make(chan struct{})

		go func() {
			c.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			logger.Debug("stopped on doneCh signal")
		case <-time.After(5 * time.Second): //nolint:mnd
			// reasonable shutdown timeout
			logger.Warn("stopped on timeout")
		}
	})
}

// Run starts the client's distribution pump.
func (c *Client) Run() {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:  c.GetHubName(),
		aichteeteapee.FieldClientID: c.id,
	})

	// Set running flag at the very start
	c.isRunning.Store(true)
	defer c.isRunning.Store(false)

	if c.hub == nil {
		logger.Error("client hub is nil, cannot run")

		return
	}

	if c.isStopped.Load() {
		logger.Debug("client is stopped, cannot run")

		return
	}

	logger.Debug("starting client distribution pump")

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		c.distributionPump()
	}()

	// Signal that we're ready to be stopped
	close(c.readyToStopCh)

	defer c.Stop()

	select {
	case <-c.doneCh:
		logger.Debug("client stopped via done channel")
	case <-c.hub.Done():
		logger.Debug("client stopped via hub.Done() channel")
	}
}

// distributionPump distributes events from sendCh to all connections.
func (c *Client) distributionPump() {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:  c.GetHubName(),
		aichteeteapee.FieldClientID: c.id,
	})

	defer func() {
		logger.Debug("distribution pump stopped")
	}()

	logger.Debug("starting client distribution pump")

	for {
		if c.isStopped.Load() {
			logger.Debug("client is done, stopping distribution pump")

			return
		}

		select {
		case event, ok := <-c.sendCh:
			if !ok {
				return // Channel closed
			}

			eventLogger := logger.WithFields(logrus.Fields{
				aichteeteapee.FieldEventType: event.Type,
				aichteeteapee.FieldEventID:   event.ID,
			})

			if c.isStopped.Load() {
				eventLogger.Debug("client is done, cannot distribute event")

				return
			}

			connections := c.GetConnections()
			if len(connections) == 0 {
				eventLogger.Debug("no connections to distribute event to")

				continue
			}

			// Distribute to all connections
			for connID, conn := range connections {
				if c.isStopped.Load() {
					logger.Debug("client done during distribution, stopping")

					return
				}

				conn.Send(event)
				eventLogger.WithField(aichteeteapee.FieldConnectionID, connID).
					Debug("event distributed to connection")
			}

		case <-c.doneCh:
			return
		}
	}
}
