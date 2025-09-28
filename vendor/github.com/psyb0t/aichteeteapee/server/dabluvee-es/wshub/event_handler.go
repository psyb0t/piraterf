package wshub

import (
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
)

// EventHandler processes events with access to the hub, client, and event,
// and returns an error if processing fails.
type EventHandler func(hub Hub, client *Client, event *dabluveees.Event) error

// EventTypeEchoRequestHandler handles echo request events by creating an
// echo reply with the same data and setting the original request event ID as
// triggeredBy.
func EventTypeEchoRequestHandler(
	_ Hub,
	client *Client,
	event *dabluveees.Event,
) error {
	// Create echo reply event with same data
	replyEvent := dabluveees.NewEvent(
		dabluveees.EventTypeEchoReply,
		event.Data,
	).SetTriggeredBy(event.ID)

	// Send the reply back to the client that sent the request
	client.SendEvent(&replyEvent)

	return nil
}
