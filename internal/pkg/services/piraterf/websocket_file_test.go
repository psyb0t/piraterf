package piraterf

import (
	"testing"

	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/common-go/env"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
}

func TestSendFileRenameSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileRenameSuccessEvent("/old/path.txt", "newname.txt")
	})
}

func TestSendFileRenameErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileRenameErrorEvent("/old/path.txt", "newname.txt", "error_type", "test error")
	})
}

func TestSendFileDeleteSuccessEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileDeleteSuccessEvent("/deleted/file.txt")
	})
}

func TestSendFileDeleteErrorEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	require.NotPanics(t, func() {
		service.sendFileDeleteErrorEvent("/deleted/file.txt", "error_type", "test error")
	})
}

func TestValidateFileRenameRequest(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	service := &PIrateRF{websocketHub: hub}

	// Test with existing fixture file
	testFile := "/workspace/.fixtures/test_2s.mp3"
	msg := fileRenameMessage{
		FilePath: testFile,
		NewName:  "newname.mp3",
	}

	require.NotPanics(t, func() {
		_, _, ok := service.validateFileRenameRequest(msg)
		assert.True(t, ok) // Basic validation should pass
	})
}