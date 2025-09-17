package piraterf

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/psyb0t/aichteeteapee"
	"github.com/psyb0t/aichteeteapee/server"
	"github.com/psyb0t/aichteeteapee/server/middleware"
	"github.com/psyb0t/aichteeteapee/server/websocket"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

func (s *PIrateRF) setupHTTPServer() error {
	if s.websocketHub == nil {
		return ctxerrors.Wrap(commonerrors.ErrNilField, "websocketHub")
	}

	httpServer, err := server.New()
	if err != nil {
		return ctxerrors.Wrap(err, "failed to create HTTP server with TLS config")
	}

	s.httpServer = httpServer

	return nil
}

func (s *PIrateRF) getHTTPServerRouter() (*server.Router, error) {
	// Ensure upload directory exists before setting up routes
	if err := s.ensureUploadDirExists(); err != nil {
		return nil, err
	}

	return &server.Router{
		GlobalMiddlewares: []middleware.Middleware{
			middleware.Recovery(),
			middleware.RequestID(),
			middleware.Logger(),
			middleware.SecurityHeaders(),
			middleware.Timeout(
				middleware.WithTimeout(time.Minute),
			),
			middleware.CORS(),
		},
		Static: []server.StaticRouteConfig{
			{
				Dir:  s.config.StaticDir,
				Path: "/static",
			},
			{
				Dir:                   s.config.FilesDir,
				Path:                  "/files",
				DirectoryIndexingType: server.DirectoryIndexingTypeJSON,
			},
		},

		Groups: []server.GroupConfig{
			{
				Path: "/",
				Routes: []server.RouteConfig{
					{
						Method:  http.MethodGet,
						Path:    "/",
						Handler: s.rootHandler,
					},
					{
						Method:  http.MethodGet,
						Path:    "/ws",
						Handler: websocket.UpgradeHandler(s.websocketHub),
					},
					{
						Method: http.MethodPost,
						Path:   "/upload",
						Handler: s.httpServer.FileUploadHandler(
							s.config.UploadDir,
							server.WithFileUploadHandlerPostprocessor(
								s.fileConversionPostprocessor,
							),
							server.WithFilenamePrependType(
								server.FilenamePrependTypeDateTime,
							),
						),
					},
				},
			},
		},
	}, nil
}

// rootHandler serves the index.html file with cache-busting timestamps.
func (s *PIrateRF) rootHandler(w http.ResponseWriter, r *http.Request) {
	// Only serve index.html for exact root path
	if r.URL.Path != "/" {
		aichteeteapee.WriteJSON(
			w,
			http.StatusNotFound,
			aichteeteapee.ErrorResponseFileNotFound,
		)

		return
	}

	// Read the HTML file
	htmlPath := path.Join(s.config.HTMLDir, aichteeteapee.FileNameIndexHTML)

	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		http.Error(w, "Failed to read HTML file", http.StatusInternalServerError)

		return
	}

	// Get file modification times for cache busting
	cssPath := filepath.Join(s.config.StaticDir, "style.css")
	jsPath := filepath.Join(s.config.StaticDir, "script.js")
	envJsPath := filepath.Join(s.config.StaticDir, "env.js")

	cssTimestamp := s.getFileTimestamp(cssPath)
	jsTimestamp := s.getFileTimestamp(jsPath)
	envJsTimestamp := s.getFileTimestamp(envJsPath)

	// Replace static URLs with timestamped versions
	htmlString := string(htmlContent)
	htmlString = strings.ReplaceAll(htmlString,
		`href="/static/style.css"`,
		fmt.Sprintf(`href="/static/style.css?v=%d"`, cssTimestamp))
	htmlString = strings.ReplaceAll(htmlString,
		`src="/static/env.js"`,
		fmt.Sprintf(`src="/static/env.js?v=%d"`, envJsTimestamp))
	htmlString = strings.ReplaceAll(htmlString,
		`src="/static/script.js"`,
		fmt.Sprintf(`src="/static/script.js?v=%d"`, jsTimestamp))

	// Set content type and serve
	w.Header().Set(
		aichteeteapee.HeaderNameContentType,
		aichteeteapee.ContentTypeHTMLUTF8,
	)
	_, _ = io.WriteString(w, htmlString)
}

// getFileTimestamp returns the Unix timestamp of the file's last modification.
func (s *PIrateRF) getFileTimestamp(filePath string) int64 {
	stat, err := os.Stat(filePath)
	if err != nil {
		logrus.WithError(err).
			Warnf("Failed to get timestamp for %s", filePath)

		return 0
	}

	return stat.ModTime().Unix()
}
