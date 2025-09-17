package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee"
)

// RequestIDMiddleware generates and injects a unique request ID
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(aichteeteapee.HeaderNameXRequestID)
			if reqID == "" {
				// Generate UUID4
				reqID = uuid.New().String()
			}

			// Set response header
			w.Header().Set(
				aichteeteapee.HeaderNameXRequestID,
				reqID,
			)

			// Add to context

			ctx := context.WithValue(
				r.Context(),
				aichteeteapee.ContextKeyRequestID,
				reqID,
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
