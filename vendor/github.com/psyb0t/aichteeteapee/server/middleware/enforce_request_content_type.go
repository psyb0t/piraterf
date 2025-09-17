package middleware

import (
	"net/http"
	"strings"

	"github.com/psyb0t/aichteeteapee"
)

// EnforceRequestContentTypeMiddleware enforces specific content types on incoming requests
func EnforceRequestContentType(allowedContentTypes ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip enforcement for GET, HEAD, DELETE methods which typically don't have request bodies
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
				next.ServeHTTP(w, r)

				return
			}

			contentType := r.Header.Get(aichteeteapee.HeaderNameContentType)
			if contentType == "" {
				aichteeteapee.WriteJSON(
					w,
					http.StatusBadRequest,
					aichteeteapee.ErrorResponseMissingContentType,
				)

				return
			}

			// Extract the media type (ignore charset and other parameters)
			mediaType := strings.Split(contentType, ";")[0]
			mediaType = strings.TrimSpace(mediaType)

			// Check if the content type is allowed
			for _, allowed := range allowedContentTypes {
				if strings.EqualFold(mediaType, allowed) {
					next.ServeHTTP(w, r)

					return
				}
			}

			// Create error response with details
			errorResponse := aichteeteapee.ErrorResponseUnsupportedContentType
			errorResponse.Details = map[string]any{
				"received": mediaType,
				"allowed":  allowedContentTypes,
			}
			aichteeteapee.WriteJSON(
				w,
				http.StatusUnsupportedMediaType,
				errorResponse,
			)
		})
	}
}

// EnforceRequestContentTypeJSONMiddleware is a convenience function that enforces JSON content type on requests
func EnforceRequestContentTypeJSON() Middleware {
	return EnforceRequestContentType(aichteeteapee.ContentTypeJSON)
}
