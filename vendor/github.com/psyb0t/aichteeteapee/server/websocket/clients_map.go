package websocket

import (
	"maps"
	"sync"

	"github.com/google/uuid"
)

type clientsMap struct {
	clients map[uuid.UUID]*Client
	mu      sync.RWMutex
}

func newClientsMap() *clientsMap {
	return &clientsMap{
		clients: make(map[uuid.UUID]*Client),
	}
}

func (cm *clientsMap) Add(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.clients[client.ID()] = client
}

func (cm *clientsMap) GetOrAdd(client *Client) (*Client, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	clientID := client.ID()
	if existingClient, exists := cm.clients[clientID]; exists {
		return existingClient, false // false means not added, just retrieved existing
	}

	cm.clients[clientID] = client

	return client, true // true means we added the client
}

func (cm *clientsMap) Remove(clientID uuid.UUID) *Client {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, exists := cm.clients[clientID]; exists {
		delete(cm.clients, clientID)

		return client
	}

	return nil
}

func (cm *clientsMap) Get(clientID uuid.UUID) *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[clientID]
	if !exists {
		return nil
	}

	return client
}

func (cm *clientsMap) GetAll() map[uuid.UUID]*Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[uuid.UUID]*Client)
	maps.Copy(result, cm.clients)

	return result
}

func (cm *clientsMap) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.clients)
}
