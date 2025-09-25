package aichteeteapee

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(
	w http.ResponseWriter,
	statusCode int,
	data any,
) {
	w.Header().Set(
		HeaderNameContentType,
		ContentTypeJSON,
	)
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty print for better readability

	if err := encoder.Encode(data); err != nil {
		logrus.WithError(err).Error("Failed to encode JSON response")
	}
}
