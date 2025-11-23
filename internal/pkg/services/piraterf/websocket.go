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

	// Preset operation handlers
	s.websocketHub.RegisterEventHandler(
		eventTypePresetLoad,
		s.handlePresetLoad,
	)

	s.websocketHub.RegisterEventHandler(
		eventTypePresetSave,
		s.handlePresetSave,
	)

	s.websocketHub.RegisterEventHandler(
		eventTypePresetRename,
		s.handlePresetRename,
	)

	s.websocketHub.RegisterEventHandler(
		eventTypePresetDelete,
		s.handlePresetDelete,
	)

	// Echo handlers (for heartbeat)
	s.websocketHub.RegisterEventHandler(
		dabluveees.EventTypeEchoRequest,
		wshub.EventTypeEchoRequestHandler,
	)
}
