package middleware

import (
	"bufio"
	"net"
	"net/http"

	"github.com/psyb0t/ctxerrors"
)

// BaseResponseWriter provides a base implementation that supports hijacking for
// WebSocket upgrades
// Other middleware can embed this to get hijacking support for free.
type BaseResponseWriter struct {
	http.ResponseWriter
}

// Hijack implements http.Hijacker interface for WebSocket support.
func (brw *BaseResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := brw.ResponseWriter.(http.Hijacker); ok {
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			return nil, nil, ctxerrors.Wrap(err, "failed to hijack connection")
		}

		return conn, rw, nil
	}

	return nil, nil, http.ErrNotSupported
}
