package piraterf

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestPath    string
		setupFiles     func(tempDir string)
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:        "serve index.html for root path",
			requestPath: "/",
			setupFiles: func(tempDir string) {
				indexPath := filepath.Join(tempDir, "index.html")
				src := ".fixtures/test_index.html"
				err := copyFile(src, indexPath)
				require.NoError(t, err)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>PIrateRF</body></html>",
			expectedHeader: "text/html; charset=UTF-8",
		},
		{
			name:        "non-root path returns 404 JSON",
			requestPath: "/style.css",
			setupFiles: func(tempDir string) {
				// Don't create the file - rootHandler only serves /
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "{\n  \"code\": \"FILE_NOT_FOUND\",\n  \"message\": \"File not found\"\n}\n",
			expectedHeader: "application/json",
		},
		{
			name:        "non-root JavaScript path returns 404 JSON",
			requestPath: "/app.js",
			setupFiles: func(tempDir string) {
				// Don't create the file - rootHandler only serves /
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "{\n  \"code\": \"FILE_NOT_FOUND\",\n  \"message\": \"File not found\"\n}\n",
			expectedHeader: "application/json",
		},
		{
			name:        "missing index.html returns 500 error",
			requestPath: "/",
			setupFiles: func(tempDir string) {
				// Don't create index.html
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to read HTML file\n",
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:        "directory traversal returns 404 JSON",
			requestPath: "/../../../etc/passwd",
			setupFiles: func(tempDir string) {
				// Create index.html but path is not /
				indexPath := filepath.Join(tempDir, "index.html")
				src := ".fixtures/test_safe.html"
				err := copyFile(src, indexPath)
				require.NoError(t, err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "{\n  \"code\": \"FILE_NOT_FOUND\",\n  \"message\": \"File not found\"\n}\n",
			expectedHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup test files
			tt.setupFiles(tempDir)

			service := &PIrateRF{
				config: Config{
					HTMLDir: tempDir,
				},
			}

			// Create HTTP request
			req := httptest.NewRequest("GET", "http://localhost"+tt.requestPath, nil)
			w := httptest.NewRecorder()

			// Call the root handler
			service.rootHandler(w, req)

			// Check response status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			// Check content type header if specified
			if tt.expectedHeader != "" {
				contentType := w.Header().Get("Content-Type")
				assert.Equal(t, tt.expectedHeader, contentType)
			}
		})
	}
}

