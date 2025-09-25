package dabluveees

import (
	"encoding/json"
	"maps"
	"sync"

	"github.com/psyb0t/ctxerrors"
)

type EventMetadataMap struct {
	data map[string]any
	mu   sync.RWMutex
}

func newEventMetadataMap() *EventMetadataMap {
	return &EventMetadataMap{
		data: make(map[string]any),
	}
}

func (emm *EventMetadataMap) Set(key string, value any) {
	emm.mu.Lock()
	defer emm.mu.Unlock()

	emm.data[key] = value
}

func (emm *EventMetadataMap) Get(key string) (any, bool) {
	emm.mu.RLock()
	defer emm.mu.RUnlock()

	value, exists := emm.data[key]

	return value, exists
}

func (emm *EventMetadataMap) GetAll() map[string]any {
	emm.mu.RLock()
	defer emm.mu.RUnlock()

	result := make(map[string]any)
	maps.Copy(result, emm.data)

	return result
}

func (emm *EventMetadataMap) Copy() *EventMetadataMap {
	emm.mu.RLock()
	defer emm.mu.RUnlock()

	newMap := newEventMetadataMap()
	maps.Copy(newMap.data, emm.data)

	return newMap
}

func (emm *EventMetadataMap) MarshalJSON() ([]byte, error) {
	emm.mu.RLock()
	defer emm.mu.RUnlock()

	data, err := json.Marshal(emm.data)
	if err != nil {
		return nil, ctxerrors.Wrap(err, "failed to marshal metadata")
	}

	return data, nil
}

func (emm *EventMetadataMap) UnmarshalJSON(data []byte) error {
	emm.mu.Lock()
	defer emm.mu.Unlock()

	if emm.data == nil {
		emm.data = make(map[string]any)
	}

	if err := json.Unmarshal(data, &emm.data); err != nil {
		return ctxerrors.Wrap(err, "failed to unmarshal metadata")
	}

	return nil
}
