package piraterf

import (
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wsunixbridge"
	"github.com/sirupsen/logrus"
)

func (s *PIrateRF) handleLiveAudioConnection(
	connection *wsunixbridge.Connection,
) error {
	logger := logrus.WithFields(logrus.Fields{
		"connectionID": connection.ID,
		"remoteAddr":   connection.Conn.RemoteAddr(),
	})

	logger.Info(
		"Live audio WebSocket connection established",
	)

	return nil
}
