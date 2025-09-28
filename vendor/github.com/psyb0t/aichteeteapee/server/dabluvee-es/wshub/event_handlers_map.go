package wshub

import (
	"sync"

	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
)

// Hub interface is implemented in hub.go

type EventHandlersMap struct {
	handlers map[dabluveees.EventType]EventHandler
	mu       sync.RWMutex
}

func NewEventHandlersMap() *EventHandlersMap {
	return &EventHandlersMap{
		handlers: make(map[dabluveees.EventType]EventHandler),
	}
}

func (ehm *EventHandlersMap) Add(
	eventType dabluveees.EventType, handler EventHandler,
) {
	ehm.mu.Lock()
	defer ehm.mu.Unlock()

	ehm.handlers[eventType] = handler
}

func (ehm *EventHandlersMap) Remove(eventType dabluveees.EventType) {
	ehm.mu.Lock()
	defer ehm.mu.Unlock()

	delete(ehm.handlers, eventType)
}

func (ehm *EventHandlersMap) Get(
	eventType dabluveees.EventType,
) (EventHandler, bool) {
	ehm.mu.RLock()
	defer ehm.mu.RUnlock()

	handler, exists := ehm.handlers[eventType]

	return handler, exists
}
