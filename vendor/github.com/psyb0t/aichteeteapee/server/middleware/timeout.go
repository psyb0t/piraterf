package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

// timeoutResponseWriter wraps http.ResponseWriter to prevent concurrent writes during timeout
type timeoutResponseWriter struct {
	BaseResponseWriter
	mu      *sync.Mutex
	written *bool
}

func (tw *timeoutResponseWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if !*tw.written {
		*tw.written = true
		tw.ResponseWriter.WriteHeader(code)
	}
}

func (tw *timeoutResponseWriter) Write(data []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	// If this is the first write, mark as written and write headers if needed
	if !*tw.written {
		*tw.written = true
	}

	// Always attempt to write the data (unless we've already timed out)
	n, err := tw.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}

	return n, nil
}

const (
	// Default timeout durations
	DefaultTimeout = 10 * time.Second
	ShortTimeout   = 5 * time.Second
	LongTimeout    = 30 * time.Second
)

// TimeoutConfig holds configuration for timeout middleware
type TimeoutConfig struct {
	Timeout time.Duration
}

type TimeoutOption func(*TimeoutConfig)

func WithTimeout(timeout time.Duration) TimeoutOption {
	return func(c *TimeoutConfig) {
		c.Timeout = timeout
	}
}

func WithDefaultTimeout() TimeoutOption {
	return func(c *TimeoutConfig) {
		c.Timeout = DefaultTimeout
	}
}

func WithShortTimeout() TimeoutOption {
	return func(c *TimeoutConfig) {
		c.Timeout = ShortTimeout
	}
}

func WithLongTimeout() TimeoutOption {
	return func(c *TimeoutConfig) {
		c.Timeout = LongTimeout
	}
}

// TimeoutMiddleware sets a timeout for the request context and handles timeout responses
//
//nolint:funlen // Timeout handling logic requires length
func Timeout(opts ...TimeoutOption) Middleware {
	config := &TimeoutConfig{
		Timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), config.Timeout)
			defer cancel()

			// Channel to track if handler completes
			doneCh := make(chan struct{})

			// Use a mutex to ensure only one response is written
			var (
				mu              sync.Mutex
				responseWritten bool
			)

			// Create a wrapper that protects against concurrent writes
			wrappedWriter := &timeoutResponseWriter{
				BaseResponseWriter: BaseResponseWriter{ResponseWriter: w},
				mu:                 &mu,
				written:            &responseWritten,
			}

			// Run the handler in a goroutine
			go func() {
				defer close(doneCh)

				next.ServeHTTP(wrappedWriter, r.WithContext(ctx))
			}()

			// Wait for either completion or timeout
			select {
			case <-doneCh:
				// Handler completed normally
				return
			case <-ctx.Done():
				// Timeout occurred - send timeout response
				mu.Lock()

				if !responseWritten {
					responseWritten = true

					logrus.WithFields(logrus.Fields{
						"method":  r.Method,
						"path":    r.URL.Path,
						"timeout": config.Timeout.String(),
					}).Info("request timeout exceeded, returning gateway timeout")

					w.Header().Set(aichteeteapee.HeaderNameContentType, aichteeteapee.ContentTypeJSON)
					w.WriteHeader(http.StatusGatewayTimeout)

					// Create a gateway timeout error response
					timeoutError := aichteeteapee.ErrorResponse{
						Code:    "GATEWAY_TIMEOUT",
						Message: "Gateway timeout - request processing took too long",
					}

					if responseBytes, err := json.Marshal(timeoutError); err == nil {
						_, _ = w.Write(responseBytes)
					} else {
						_, _ = w.Write([]byte(`{"code":"GATEWAY_TIMEOUT","message":"Gateway timeout"}`))
					}
				}

				mu.Unlock()
			}
		})
	}
}
