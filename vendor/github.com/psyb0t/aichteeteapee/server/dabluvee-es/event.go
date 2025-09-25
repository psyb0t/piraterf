package dabluveees

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

type EventType string

const (
	EventTypeSystemLog   EventType = "system.log"
	EventTypeShellExec   EventType = "shell.exec"
	EventTypeEchoRequest EventType = "echo.request"
	EventTypeEchoReply   EventType = "echo.reply"
	EventTypeError       EventType = "error"
)

type Event struct {
	ID   uuid.UUID       `json:"id"` // UUID4 identifier
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
	// Unix timestamp (seconds) - SET BY SENDER
	Timestamp int64             `json:"timestamp"`
	Metadata  *EventMetadataMap `json:"metadata"` // For rooms, userID, etc.
}

// NewEvent creates a new event with current unix timestamp
// Use this when the SERVER is creating/sending an event.
func NewEvent(eventType EventType, data any) *Event {
	eventID := uuid.New()

	logger := logrus.WithFields(logrus.Fields{
		aichteeteapee.FieldEventID:   eventID,
		aichteeteapee.FieldEventType: string(eventType),
	})

	logger.Debug("creating new event")

	var rawData json.RawMessage
	if data != nil {
		if jsonData, err := json.Marshal(data); err != nil {
			logger.WithError(err).Error("failed to marshal event data, using nil")
		} else {
			rawData = jsonData
		}
	}

	return &Event{
		ID:        eventID,
		Type:      eventType,
		Data:      rawData,
		Timestamp: time.Now().Unix(), // Server sets timestamp when server sends
		Metadata:  newEventMetadataMap(),
	}
}

// WithMetadata adds metadata to an event (chainable).
func (e Event) WithMetadata(key string, value any) Event {
	if e.Metadata == nil {
		e.Metadata = newEventMetadataMap()
	}

	e.Metadata.Set(key, value)

	return e
}

// WithTimestamp sets a specific unix timestamp (chainable).
func (e Event) WithTimestamp(unixTimestamp int64) Event {
	e.Timestamp = unixTimestamp

	return e
}

// GetTime converts unix timestamp to time.Time for Go usage.
func (e Event) GetTime() time.Time {
	return time.Unix(e.Timestamp, 0)
}

// IsRecent checks if event is within the last N seconds.
func (e Event) IsRecent(seconds int64) bool {
	return time.Now().Unix()-e.Timestamp <= seconds
}
