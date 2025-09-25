package wshub

import (
	"sync"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/sirupsen/logrus"
)

// Hub interface needs all methods for complete functionality
//
//nolint:interfacebloat
type Hub interface {
	Name() string
	Close()
	AddClient(client *Client)
	RemoveClient(clientID uuid.UUID)
	GetClient(clientID uuid.UUID) *Client
	GetOrCreateClient(clientID uuid.UUID, opts ...ClientOption) (*Client, bool)
	GetAllClients() map[uuid.UUID]*Client
	RegisterEventHandler(eventType dabluveees.EventType, handler EventHandler)
	RegisterEventHandlers(handlers map[dabluveees.EventType]EventHandler)
	UnregisterEventHandler(eventType dabluveees.EventType)
	ProcessEvent(client *Client, event *dabluveees.Event)
	BroadcastToAll(event *dabluveees.Event)
	BroadcastToClients(clientIDs []uuid.UUID, event *dabluveees.Event)
	BroadcastToSubscribers(eventType dabluveees.EventType, event *dabluveees.Event)
	Done() <-chan struct{}
}

type hub struct {
	name     string
	clients  *clientsMap
	handlers *EventHandlersMap
	doneCh   chan struct{}
	wg       sync.WaitGroup
	stopOnce sync.Once
	mu       sync.Mutex
}

func NewHub(name string) Hub { //nolint:ireturn
	logger := logrus.WithField(aichteeteapee.FieldHubName, name)
	logger.Info("creating new hub")

	return &hub{
		name:     name,
		clients:  newClientsMap(),
		handlers: NewEventHandlersMap(),
		doneCh:   make(chan struct{}),
	}
}

func (h *hub) Name() string {
	return h.name
}

func (h *hub) Close() {
	logger := logrus.WithField(aichteeteapee.FieldHubName, h.name)

	h.mu.Lock()
	defer h.mu.Unlock()

	h.stopOnce.Do(func() {
		logger.Info("closing hub")

		close(h.doneCh)

		// Stop all clients
		clients := h.clients.GetAll()
		logger.WithField(aichteeteapee.FieldTotalClients, len(clients)).
			Debug("stopping all hub clients")

		for _, client := range clients {
			go client.Stop()
		}

		h.wg.Wait()

		// Clear all clients from the hub after stopping them
		for clientID := range clients {
			h.clients.Remove(clientID)
		}
	})
}

func (h *hub) Done() <-chan struct{} {
	return h.doneCh
}

func (h *hub) AddClient(client *Client) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:  h.name,
		aichteeteapee.FieldClientID: client.id,
	})

	h.mu.Lock()
	defer h.mu.Unlock()

	logger.Debug("adding client to hub")

	// Inject hub reference and doneCh into client
	client.SetHub(h)

	h.clients.Add(client)

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()
		defer func() {
			// Remove client from hub when Run() exits
			logger.Debug("client run finished, removing from hub")
			h.clients.Remove(client.id)
		}()

		client.Run()
	}()
}

func (h *hub) RemoveClient(clientID uuid.UUID) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:  h.name,
		aichteeteapee.FieldClientID: clientID,
	})

	logger.Debug("removing client from hub")

	if client := h.clients.Remove(clientID); client != nil {
		client.Stop()
	}
}

func (h *hub) GetClient(clientID uuid.UUID) *Client {
	return h.clients.Get(clientID)
}

func (h *hub) GetOrCreateClient(
	clientID uuid.UUID, opts ...ClientOption,
) (*Client, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:  h.name,
		aichteeteapee.FieldClientID: clientID,
	})

	// Check if client already exists
	if existingClient := h.clients.Get(clientID); existingClient != nil {
		logger.Debug("client already exists in hub")

		return existingClient, false // false means not created, just retrieved
	}

	// Create new client and add it to hub properly
	logger.Debug("creating new client and adding to hub")

	newClient := NewClientWithID(clientID, opts...)

	// Set up the client properly (same as AddClient but without
	// the lock since we already have it)
	newClient.SetHub(h)
	h.clients.Add(newClient)

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()
		defer func() {
			// Remove client from hub when Run() exits
			logger.Debug("client run finished, removing from hub")
			h.clients.Remove(newClient.id)
		}()

		newClient.Run()
	}()

	return newClient, true // true means we created it
}

func (h *hub) GetAllClients() map[uuid.UUID]*Client {
	return h.clients.GetAll()
}

func (h *hub) RegisterEventHandler(
	eventType dabluveees.EventType, handler EventHandler,
) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:   h.name,
		aichteeteapee.FieldEventType: string(eventType),
	})

	logger.Info("registering event handler for event")

	h.handlers.Add(eventType, handler)
}

func (h *hub) RegisterEventHandlers(
	handlers map[dabluveees.EventType]EventHandler,
) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName: h.name,
	})

	logger.WithField("count", len(handlers)).
		Info("registering multiple event handlers")

	for eventType, handler := range handlers {
		h.RegisterEventHandler(eventType, handler)
	}
}

func (h *hub) UnregisterEventHandler(eventType dabluveees.EventType) {
	h.handlers.Remove(eventType)
}

func (h *hub) ProcessEvent(client *Client, event *dabluveees.Event) {
	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldHubName:   h.name,
		aichteeteapee.FieldClientID:  client.ID(),
		aichteeteapee.FieldEventType: string(event.Type),
		aichteeteapee.FieldEventID:   event.ID,
	})

	logger.Debug("processing event through hub")

	handler, exists := h.handlers.Get(event.Type)
	if !exists {
		logger.Debug("no handler registered for event type")

		return
	}

	if err := handler(h, client, event); err != nil {
		logger.WithError(err).Error("event handler execution failed")
		// Note: Don't return error here as per v4 spec - just log it
	}
}

func (h *hub) BroadcastToAll(event *dabluveees.Event) {
	clients := h.clients.GetAll()
	for _, client := range clients {
		client.SendEvent(event)
	}
}

func (h *hub) BroadcastToClients(
	clientIDs []uuid.UUID, event *dabluveees.Event,
) {
	for _, clientID := range clientIDs {
		client := h.clients.Get(clientID)
		if client != nil {
			client.SendEvent(event)
		}
	}
}

func (h *hub) BroadcastToSubscribers(
	eventType dabluveees.EventType, event *dabluveees.Event,
) {
	clients := h.clients.GetAll()
	for _, client := range clients {
		if client.IsSubscribedTo(eventType) {
			client.SendEvent(event)
		}
	}
}
