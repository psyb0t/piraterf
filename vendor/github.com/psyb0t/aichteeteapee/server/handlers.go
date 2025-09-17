package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee"
)

// HealthHandler provides a basic health check endpoint
func (s *Server) HealthHandler(
	w http.ResponseWriter,
	_ *http.Request,
) {
	response := map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	}

	aichteeteapee.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}

// EchoHandler echoes back request information (useful for testing)
func (s *Server) EchoHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Check if request has a body and if so, ensure it's JSON
	if r.Body != nil && r.ContentLength > 0 {
		if !aichteeteapee.IsRequestContentTypeJSON(r) {
			aichteeteapee.WriteJSON(
				w,
				http.StatusUnsupportedMediaType,
				aichteeteapee.ErrorResponseUnsupportedContentType,
			)

			return
		}
	}

	var body any

	if r.Body != nil {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&body); err != nil {
			s.logger.WithError(err).Error("Failed to decode request body in echo handler")
		}
	}

	response := map[string]any{
		"method":  r.Method,
		"path":    r.URL.Path,
		"query":   r.URL.Query(),
		"headers": r.Header,
		"body":    body,
	}

	// Add user if available (from auth middleware)
	if user, ok := r.Context().Value(
		aichteeteapee.ContextKeyUser,
	).(string); ok {
		response["user"] = user
	}

	aichteeteapee.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}

type FilenamePrependType uint8

const (
	// FilenamePrependTypeNone does not add any prefix to the filename
	FilenamePrependTypeNone FilenamePrependType = iota
	// FilenamePrependTypeDateTime prepends date and time in Y_M_D_H_I_S format
	FilenamePrependTypeDateTime
	// FilenamePrependTypeUUID prepends a UUID4 to the filename (default)
	FilenamePrependTypeUUID
)

type FileUploadHandlerOption func(*FileUploadHandlerConfig)

// FileUploadHandlerConfig holds the configuration for the file upload handler
type FileUploadHandlerConfig struct {
	postprocessor   func(map[string]any) (map[string]any, error)
	filenamePrepend FilenamePrependType
}

// WithFileUploadHandlerPostprocessor sets a postprocessor function that modifies the response
func WithFileUploadHandlerPostprocessor(
	fn func(map[string]any) (map[string]any, error),
) FileUploadHandlerOption {
	return func(config *FileUploadHandlerConfig) {
		config.postprocessor = fn
	}
}

// WithFilenamePrependType sets the type of prefix to add to uploaded filenames
func WithFilenamePrependType(
	prependType FilenamePrependType,
) FileUploadHandlerOption {
	return func(config *FileUploadHandlerConfig) {
		config.filenamePrepend = prependType
	}
}

// FileUploadHandler returns a handler for file uploads to the specified directory
func (s *Server) FileUploadHandler(
	uploadsDir string,
	opts ...FileUploadHandlerOption,
) http.HandlerFunc {
	// Apply options
	config := &FileUploadHandlerConfig{
		filenamePrepend: FilenamePrependTypeUUID, // default to UUID
	}
	for _, opt := range opts {
		opt(config)
	}

	// Ensure the uploads directory exists
	const dirPermissions = 0o750
	if err := os.MkdirAll(uploadsDir, dirPermissions); err != nil {
		s.logger.WithError(err).
			WithField("dir", uploadsDir).
			Error("Failed to create uploads directory")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			aichteeteapee.WriteJSON(
				w,
				http.StatusMethodNotAllowed,
				aichteeteapee.ErrorResponseMethodNotAllowed,
			)

			return
		}

		if err := s.handleFileUpload(w, r, uploadsDir, config); err != nil {
			// Error already handled in handleFileUpload
			return
		}
	}
}

// handleFileUpload processes the file upload and writes the response
//
//nolint:funlen // Complex file upload logic requires length
func (s *Server) handleFileUpload(
	w http.ResponseWriter,
	r *http.Request,
	uploadsDir string,
	config *FileUploadHandlerConfig,
) error {
	if err := r.ParseMultipartForm(s.config.FileUploadMaxMemory); err != nil {
		s.logger.WithError(err).Error("Failed to parse multipart form")
		aichteeteapee.WriteJSON(
			w,
			http.StatusBadRequest,
			aichteeteapee.ErrorResponseInvalidMultipartForm,
		)

		return fmt.Errorf("parse multipart form: %w", err)
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		s.logger.WithError(err).Error("Failed to get file from form")
		aichteeteapee.WriteJSON(
			w,
			http.StatusBadRequest,
			aichteeteapee.ErrorResponseNoFileProvided,
		)

		return fmt.Errorf("get form file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			s.logger.WithError(closeErr).
				Error("Failed to close uploaded file")
		}
	}()

	// Generate filename with configured prepend type
	uniqueFilename := s.generateUniqueFilename(handler.Filename, config.filenamePrepend)
	filePath := filepath.Join(uploadsDir, uniqueFilename)

	if err := s.saveUploadedFile(file, filePath); err != nil {
		aichteeteapee.WriteJSON(
			w,
			http.StatusInternalServerError,
			aichteeteapee.ErrorResponseFileSaveFailed,
		)

		return err
	}

	// Get absolute path for response
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		s.logger.WithError(err).
			WithField("path", filePath).
			Warn("Failed to get absolute path, using relative path")

		absolutePath = filePath
	}

	response := map[string]any{
		"status":            "success",
		"original_filename": handler.Filename,
		"saved_filename":    uniqueFilename,
		"size":              handler.Size,
		"path":              absolutePath,
	}

	// Apply postprocessor if configured
	if config.postprocessor != nil {
		processedResponse, err := config.postprocessor(response)
		if err != nil {
			s.logger.WithError(err).Error("Failed to postprocess response")
			aichteeteapee.WriteJSON(
				w,
				http.StatusInternalServerError,
				aichteeteapee.ErrorResponseInternalServerError,
			)

			return fmt.Errorf("postprocess response: %w", err)
		}

		response = processedResponse
	}

	aichteeteapee.WriteJSON(
		w,
		http.StatusOK,
		response,
	)

	return nil
}

// saveUploadedFile saves the uploaded file to the specified path
func (s *Server) saveUploadedFile(
	src io.Reader,
	filePath string,
) error {
	dst, err := os.Create(filePath)
	if err != nil {
		s.logger.WithError(err).
			WithField("path", filePath).
			Error("Failed to create destination file")

		return fmt.Errorf("create file %s: %w", filePath, err)
	}

	defer func() {
		if closeErr := dst.Close(); closeErr != nil {
			s.logger.WithError(closeErr).
				Error("Failed to close destination file")
		}
	}()

	if _, err := io.Copy(dst, src); err != nil {
		s.logger.WithError(err).Error("Failed to copy file content")

		return fmt.Errorf("copy file content: %w", err)
	}

	return nil
}

// generateUniqueFilename creates a unique filename based on the prepend type
func (s *Server) generateUniqueFilename(
	originalFilename string,
	prependType FilenamePrependType,
) string {
	switch prependType {
	case FilenamePrependTypeNone:
		return originalFilename
	case FilenamePrependTypeDateTime:
		now := time.Now()
		dateTimePrefix := now.Format("2006_01_02_15_04_05")

		return fmt.Sprintf("%s_%s", dateTimePrefix, originalFilename)
	case FilenamePrependTypeUUID:
		fallthrough
	default:
		uniqueID := uuid.New().String()

		return fmt.Sprintf("%s_%s", uniqueID, originalFilename)
	}
}
