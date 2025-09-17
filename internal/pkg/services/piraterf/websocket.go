package piraterf

import (
	"github.com/psyb0t/aichteeteapee/server/websocket"
)

func (s *PIrateRF) setupWebsocketHub() {
	s.websocketHub = websocket.NewHub("piraterf")

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
}
