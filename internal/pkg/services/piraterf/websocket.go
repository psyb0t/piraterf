package piraterf

import (
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
)

func (s *PIrateRF) setupWebsocketHub() {
	s.websocketHub = wshub.NewHub("piraterf")

	// RPITX execution handlers
	s.websocketHub.RegisterEventHandler(
		eventTypeRPITXExecutionStart,
		s.handleRPITXExecutionStart,
	)

	s.websocketHub.RegisterEventHandler(
		eventTypeRPITXExecutionStop,
		s.handleRPITXExecutionStop,
	)

	// File operation handlers
	s.websocketHub.RegisterEventHandler(
		eventTypeFileRename,
		s.handleFileRename,
	)

	s.websocketHub.RegisterEventHandler(
		eventTypeFileDelete,
		s.handleFileDelete,
	)

	// Audio operation handlers
	s.websocketHub.RegisterEventHandler(
		eventTypeAudioPlaylistCreate,
		s.handleAudioPlaylistCreate,
	)

	// Echo handlers (for heartbeat)
	s.websocketHub.RegisterEventHandler(
		dabluveees.EventTypeEchoRequest,
		wshub.EventTypeEchoRequestHandler,
	)
}
