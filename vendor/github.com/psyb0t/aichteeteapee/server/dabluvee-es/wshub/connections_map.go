package wshub

import (
	"maps"
	"sync"

	"github.com/google/uuid"
)

// connectionsMap manages connection storage with thread-safe operations.
type connectionsMap struct {
	conns map[uuid.UUID]*Connection
	mu    sync.RWMutex
}

func newConnectionsMap() *connectionsMap {
	return &connectionsMap{
		conns: make(map[uuid.UUID]*Connection),
	}
}

func (cm *connectionsMap) Add(conn *Connection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.conns[conn.id] = conn
}

func (cm *connectionsMap) Remove(
	connectionID uuid.UUID,
) *Connection {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.conns[connectionID]; exists {
		delete(cm.conns, connectionID)

		return conn
	}

	return nil
}

// Get retrieves a connection by ID.
func (cm *connectionsMap) Get(connectionID uuid.UUID) *Connection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.conns[connectionID]
	if !exists {
		return nil
	}

	return conn
}

func (cm *connectionsMap) GetAll() map[uuid.UUID]*Connection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[uuid.UUID]*Connection)
	maps.Copy(result, cm.conns)

	return result
}

func (cm *connectionsMap) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.conns)
}

func (cm *connectionsMap) IsEmpty() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.conns) == 0
}
