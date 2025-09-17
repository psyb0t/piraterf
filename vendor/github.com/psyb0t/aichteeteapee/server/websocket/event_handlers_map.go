package websocket

import (
	"sync"
)

// EventHandler processes events with access to the hub, client, and event, and returns an error if processing fails
type EventHandler func(hub Hub, client *Client, event *Event) error

// Hub interface is implemented in hub.go

type eventHandlersMap struct {
	handlers map[EventType]EventHandler
	mu       sync.RWMutex
}

func newEventHandlersMap() *eventHandlersMap {
	return &eventHandlersMap{
		handlers: make(map[EventType]EventHandler),
	}
}

func (ehm *eventHandlersMap) Add(eventType EventType, handler EventHandler) {
	ehm.mu.Lock()
	defer ehm.mu.Unlock()

	ehm.handlers[eventType] = handler
}

func (ehm *eventHandlersMap) Remove(eventType EventType) {
	ehm.mu.Lock()
	defer ehm.mu.Unlock()

	delete(ehm.handlers, eventType)
}

func (ehm *eventHandlersMap) Get(eventType EventType) (EventHandler, bool) {
	ehm.mu.RLock()
	defer ehm.mu.RUnlock()

	handler, exists := ehm.handlers[eventType]

	return handler, exists
}
