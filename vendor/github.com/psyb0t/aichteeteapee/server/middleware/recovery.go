package middleware

import (
	"encoding/json"
	"maps"
	"net/http"
	"runtime/debug"

	"github.com/psyb0t/aichteeteapee"
	"github.com/sirupsen/logrus"
)

// RecoveryConfig holds configuration for recovery middleware.
type RecoveryConfig struct {
	Logger        *logrus.Logger
	LogLevel      logrus.Level
	LogMessage    string
	StatusCode    int
	Response      any
	ContentType   string
	IncludeStack  bool
	ExtraFields   map[string]any
	CustomHandler func(recovered any, w http.ResponseWriter, r *http.Request)
}

type RecoveryOption func(*RecoveryConfig)

// WithRecoveryLogger sets the logger instance.
func WithRecoveryLogger(logger *logrus.Logger) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.Logger = logger
	}
}

// WithRecoveryLogLevel sets the log level for panic recovery.
func WithRecoveryLogLevel(level logrus.Level) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.LogLevel = level
	}
}

// WithRecoveryLogMessage sets the log message for panic recovery.
func WithRecoveryLogMessage(message string) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.LogMessage = message
	}
}

// WithRecoveryStatusCode sets the HTTP status code for panic responses.
func WithRecoveryStatusCode(code int) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.StatusCode = code
	}
}

// WithRecoveryResponse sets the response body for panic responses.
func WithRecoveryResponse(response any) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.Response = response
	}
}

// WithRecoveryContentType sets the content type for panic responses.
func WithRecoveryContentType(contentType string) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.ContentType = contentType
	}
}

// WithIncludeStack enables/disables stack trace inclusion in logs.
func WithIncludeStack(include bool) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.IncludeStack = include
	}
}

// WithRecoveryExtraFields adds extra fields to panic log entries.
func WithRecoveryExtraFields(fields map[string]any) RecoveryOption {
	return func(c *RecoveryConfig) {
		if c.ExtraFields == nil {
			c.ExtraFields = make(map[string]any)
		}

		maps.Copy(c.ExtraFields, fields)
	}
}

// WithCustomRecoveryHandler sets a custom handler for panic recovery.
func WithCustomRecoveryHandler(
	handler func(recovered any, w http.ResponseWriter, r *http.Request),
) RecoveryOption {
	return func(c *RecoveryConfig) {
		c.CustomHandler = handler
	}
}

// RecoveryMiddleware recovers from panics with configurable options
//
// Complex panic handling logic is necessary for proper recovery
//
//nolint:gocognit,nestif,cyclop,funlen
func Recovery(opts ...RecoveryOption) Middleware {
	config := &RecoveryConfig{
		Logger:        logrus.StandardLogger(),
		LogLevel:      logrus.ErrorLevel,
		LogMessage:    "Panic recovered in HTTP handler",
		StatusCode:    http.StatusInternalServerError,
		Response:      aichteeteapee.ErrorResponseInternalServerError,
		ContentType:   aichteeteapee.ContentTypeJSON,
		IncludeStack:  true, // Enable stack traces by default
		ExtraFields:   make(map[string]any),
		CustomHandler: nil,
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					// Use custom handler if provided
					if config.CustomHandler != nil {
						config.CustomHandler(recovered, w, r)

						return
					}

					reqID := aichteeteapee.GetRequestID(r)

					// Build log fields
					fields := logrus.Fields{
						"error":     recovered,
						"method":    r.Method,
						"path":      r.URL.Path,
						"ip":        aichteeteapee.GetClientIP(r),
						"requestId": reqID,
					}

					// Add extra fields
					maps.Copy(fields, config.ExtraFields)

					if config.IncludeStack {
						fields["stack"] = string(debug.Stack())
					}

					config.Logger.WithFields(fields).Log(config.LogLevel, config.LogMessage)

					// Set content type if not already set
					if w.Header().Get(aichteeteapee.HeaderNameContentType) == "" {
						w.Header().Set(aichteeteapee.HeaderNameContentType, config.ContentType)
					}

					w.WriteHeader(config.StatusCode)

					// Handle JSON response encoding safely
					if config.ContentType == aichteeteapee.ContentTypeJSON {
						// Try to encode the response
						jsonData, err := json.Marshal(config.Response)
						if err != nil {
							// If encoding fails, use a hardcoded fallback
							config.Logger.Errorf(
								"Failed to encode error response during panic recovery: %v", err,
							)

							fallbackResponse := []byte(
								`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error"}`,
							)
							if _, writeErr := w.Write(fallbackResponse); writeErr != nil {
								config.Logger.Errorf(
									"Failed to write fallback response during panic recovery: %v",
									writeErr,
								)
							}
						} else {
							// Encoding succeeded, write the response
							if _, writeErr := w.Write(jsonData); writeErr != nil {
								config.Logger.Errorf(
									"Failed to write JSON response during panic recovery: %v", writeErr,
								)
							}
						}
					} else {
						// For non-JSON responses, try to write as string
						if str, ok := config.Response.(string); ok {
							if _, err := w.Write([]byte(str)); err != nil {
								config.Logger.Errorf(
									"Failed to write error response during panic recovery: %v", err,
								)
							}
						}
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
