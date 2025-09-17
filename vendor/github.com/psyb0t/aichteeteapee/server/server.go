package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/psyb0t/aichteeteapee"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

const (
	// File cache configuration constants
	defaultFileCacheTTL     = 30 * time.Second // Cache entries for 30 seconds
	defaultFileCacheMaxSize = 1000             // Maximum 1000 cached entries
)

type fileCacheEntry struct {
	exists    bool
	isDir     bool
	timestamp time.Time
}

// fileCache provides a simple cache for file existence checks to prevent DoS
type fileCache struct {
	cache   map[string]fileCacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	maxSize int
}

func newFileCache() *fileCache {
	return &fileCache{
		cache:   make(map[string]fileCacheEntry),
		ttl:     defaultFileCacheTTL,
		maxSize: defaultFileCacheMaxSize,
	}
}

func (fc *fileCache) get(
	path string,
) (bool, bool, bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	entry, exists := fc.cache[path]
	if !exists {
		return false, false, false
	}

	// Check if entry is expired
	if time.Since(entry.timestamp) > fc.ttl {
		// Clean up expired entry
		delete(fc.cache, path)

		return false, false, false
	}

	return true, entry.exists, entry.isDir
}

func (fc *fileCache) set(
	path string,
	exists bool,
	isDir bool,
) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// If cache is full, clear it (simple eviction)
	if len(fc.cache) >= fc.maxSize {
		fc.cache = make(map[string]fileCacheEntry)
	}

	fc.cache[path] = fileCacheEntry{
		exists:    exists,
		isDir:     isDir,
		timestamp: time.Now(),
	}
}

type Server struct {
	httpServer    *http.Server
	httpsServer   *http.Server
	mux           *http.ServeMux
	logger        *logrus.Logger
	config        Config
	router        *Router
	httpListener  net.Listener
	httpsListener net.Listener
	listenerMu    sync.RWMutex
	doneCh        chan struct{}
	stopOnce      sync.Once
	rootGroupOnce sync.Once
	fileCache     *fileCache
}

func New() (*Server, error) {
	return NewWithLogger(logrus.StandardLogger())
}

func NewWithLogger(
	logger *logrus.Logger,
) (*Server, error) {
	cfg, err := parseConfig()
	if err != nil {
		return nil, ctxerrors.Wrap(err, "failed to parse config")
	}

	return NewWithConfigAndLogger(cfg, logger)
}

func NewWithConfig(
	config Config,
) (*Server, error) {
	return NewWithConfigAndLogger(config, logrus.StandardLogger())
}

func NewWithConfigAndLogger(
	config Config,
	logger *logrus.Logger,
) (*Server, error) {
	mux := http.NewServeMux()

	// Create HTTP server with secure defaults
	httpServer := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           mux,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	// Create HTTPS server with same config but different address
	httpsServer := &http.Server{
		Addr:              config.TLSListenAddress,
		Handler:           mux,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	server := &Server{
		httpServer:  httpServer,
		httpsServer: httpsServer,
		mux:         mux,
		logger:      logger,
		config:      config,
		router:      &Router{}, // Initialize empty router
		doneCh:      make(chan struct{}),
		fileCache:   newFileCache(),
	}

	return server, nil
}

func (s *Server) setupGlobalMiddlewares() {
	root := NewGroup(
		s.mux,
		"/",
		s.logger,
	)

	// Apply configured global middlewares directly
	root.Use(s.router.GlobalMiddlewares...)
	s.router.rootGroup = root
}

func (s *Server) setupRoutes() {
	// Ensure we have a root group to work with
	rootGroup := s.GetRootGroup()

	// Setup route groups from router
	for _, groupConfig := range s.router.Groups {
		s.setupRouteGroup(groupConfig, rootGroup)
	}

	// Setup static routes if configured
	for _, staticConfig := range s.router.Static {
		if staticConfig.Dir != "" && staticConfig.Path != "" {
			s.setupStaticRoute(staticConfig)
		}
	}
}

func (s *Server) setupRouteGroup(
	groupConfig GroupConfig,
	parentGroup *Group,
) {
	group := parentGroup.Group(groupConfig.Path)

	// Apply group-specific middlewares directly
	group.Use(groupConfig.Middlewares...)

	// Setup routes for this group
	for _, routeConfig := range groupConfig.Routes {
		s.setupRoute(group, routeConfig)
	}

	// Setup nested groups
	for _, nestedGroupConfig := range groupConfig.Groups {
		s.setupRouteGroup(nestedGroupConfig, group)
	}
}

func (s *Server) setupRoute(
	group *Group,
	routeConfig RouteConfig,
) {
	if routeConfig.Handler == nil {
		s.logger.Warnf(
			"Route %s %s has no handler",
			routeConfig.Method,
			routeConfig.Path,
		)

		return
	}

	group.HandleFunc(
		routeConfig.Method,
		routeConfig.Path,
		routeConfig.Handler,
	)
}

func (s *Server) setupStaticRoute(
	staticConfig StaticRouteConfig,
) {
	// Ensure we have a root group to work with
	rootGroup := s.GetRootGroup()

	// Use the cleaner Handle method with proper path pattern
	staticPattern := staticConfig.Path
	if !strings.HasSuffix(staticPattern, "/") {
		staticPattern += "/"
	}

	staticPattern += "{path...}"

	// Create a secure static file handler with configurable directory listing
	handler := s.createSecureStaticHandler(staticConfig)
	rootGroup.Handle(
		http.MethodGet,
		staticPattern,
		handler,
	)
}

// createSecureStaticHandler creates a static file handler with configurable directory listing
func (s *Server) createSecureStaticHandler(
	staticConfig StaticRouteConfig,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath, ok := s.validateAndBuildPath(w, r, staticConfig.Dir, staticConfig.Path)
		if !ok {
			return
		}

		handled := s.checkFileAccess(w, r, fullPath, staticConfig)
		if !handled {
			return
		}

		// If we get here, it's a regular file (not a directory), serve it
		http.ServeFile(w, r, fullPath)
	})
}

// validateAndBuildPath validates the request path and builds the full file path
func (s *Server) validateAndBuildPath(
	w http.ResponseWriter,
	r *http.Request,
	staticDir, staticPath string,
) (string, bool) {
	// Remove the static path prefix
	relativePath := strings.TrimPrefix(r.URL.Path, staticPath)
	if relativePath == "" {
		relativePath = "."
	}

	// Security: prevent directory traversal by cleaning the path
	cleanPath := filepath.Clean(relativePath)
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, ".."+string(filepath.Separator)) {
		aichteeteapee.WriteJSON(
			w,
			http.StatusForbidden,
			aichteeteapee.ErrorResponsePathTraversalDenied,
		)

		return "", false
	}

	// Build the full file path using proper path joining
	return filepath.Join(staticDir, cleanPath), true
}

// checkFileAccess checks file existence and permissions using cache
// Returns true if it's a regular file that should be served by the caller
// Returns false if the request was handled (error, directory, etc) or should not continue
func (s *Server) checkFileAccess(
	w http.ResponseWriter,
	r *http.Request,
	fullPath string,
	staticConfig StaticRouteConfig,
) bool {
	cached, exists, isDir := s.fileCache.get(fullPath)
	if cached {
		return s.handleCachedResult(w, r, fullPath, exists, isDir, staticConfig)
	}

	return s.handleCacheMiss(w, r, fullPath, staticConfig)
}

// handleCachedResult processes cached file information
// Returns true only if it's a regular file that should be served by the caller
func (s *Server) handleCachedResult(
	w http.ResponseWriter,
	r *http.Request,
	fullPath string,
	exists, isDir bool,
	staticConfig StaticRouteConfig,
) bool {
	if !exists {
		aichteeteapee.WriteJSON(
			w,
			http.StatusNotFound,
			aichteeteapee.ErrorResponseFileNotFound,
		)

		return false
	}

	if isDir {
		s.handleDirectoryRequest(w, r, fullPath, staticConfig)

		return false // Directory handled, don't continue
	}

	return true // Regular file, should be served by caller
}

// handleCacheMiss processes filesystem lookup and caches the result
func (s *Server) handleCacheMiss(
	w http.ResponseWriter,
	r *http.Request,
	fullPath string,
	staticConfig StaticRouteConfig,
) bool {
	info, err := os.Stat(fullPath)
	if err != nil {
		// File doesn't exist - cache this result
		s.fileCache.set(fullPath, false, false)
		aichteeteapee.WriteJSON(
			w,
			http.StatusNotFound,
			aichteeteapee.ErrorResponseFileNotFound,
		)

		return false
	}

	// File exists - cache this result
	isDir := info.IsDir()
	s.fileCache.set(fullPath, true, isDir)

	if isDir {
		s.handleDirectoryRequest(w, r, fullPath, staticConfig)

		return false // Directory handled, don't continue
	}

	return true // Regular file, should be served by caller
}

// validateTLSConfig validates TLS configuration when TLS is enabled
func (s *Server) validateTLSConfig() error {
	// If TLS is not enabled, no validation needed
	if !s.config.TLSEnabled {
		return nil
	}

	if s.config.TLSCertFile == "" {
		return ctxerrors.Wrap(
			commonerrors.ErrTLSCertFileNotSpecified,
			"TLS enabled but no cert file provided",
		)
	}

	if s.config.TLSKeyFile == "" {
		return ctxerrors.Wrap(
			commonerrors.ErrTLSKeyFileNotSpecified,
			"TLS enabled but no key file provided",
		)
	}

	// Validate cert and key files exist and are readable
	if _, err := os.Stat(s.config.TLSCertFile); err != nil {
		return ctxerrors.Wrapf(err, "TLS cert file not accessible: %s", s.config.TLSCertFile)
	}

	if _, err := os.Stat(s.config.TLSKeyFile); err != nil {
		return ctxerrors.Wrapf(err, "TLS key file not accessible: %s", s.config.TLSKeyFile)
	}

	return nil
}

// Start starts the HTTP server and blocks until context is cancelled or Stop is called
func (s *Server) Start(
	ctx context.Context,
	router *Router,
) error {
	if err := s.setupRouter(router); err != nil {
		return err
	}

	defer func() {
		if stopErr := s.Stop(ctx); stopErr != nil {
			s.logger.Errorf("Error during cleanup: %v", stopErr)
		}
	}()

	httpListener, httpsListener, err := s.createListeners(ctx)
	if err != nil {
		return err
	}

	return s.serveAndWait(ctx, httpListener, httpsListener)
}

func (s *Server) setupRouter(
	router *Router,
) error {
	s.router = router

	if router != nil && len(router.GlobalMiddlewares) > 0 {
		s.setupGlobalMiddlewares()
	}

	if router != nil {
		s.setupRoutes()
	}

	return s.validateTLSConfig()
}

func (s *Server) createListeners(
	ctx context.Context,
) (net.Listener, net.Listener, error) {
	lc := &net.ListenConfig{}

	httpListener, err := lc.Listen(ctx, "tcp", s.httpServer.Addr)
	if err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to create HTTP listener")
	}

	s.listenerMu.Lock()
	s.httpListener = httpListener
	s.listenerMu.Unlock()

	var httpsListener net.Listener
	if s.config.TLSEnabled {
		httpsListener, err = lc.Listen(ctx, "tcp", s.httpsServer.Addr)
		if err != nil {
			if closeErr := httpListener.Close(); closeErr != nil {
				s.logger.Errorf("Failed to close HTTP listener: %v", closeErr)
			}

			return nil, nil, ctxerrors.Wrap(err, "failed to create HTTPS listener")
		}

		s.listenerMu.Lock()
		s.httpsListener = httpsListener
		s.listenerMu.Unlock()
	}

	return httpListener, httpsListener, nil
}

func (s *Server) serveAndWait(
	ctx context.Context,
	httpListener, httpsListener net.Listener,
) error {
	s.logServerStart()

	httpServerErrCh, httpsServerErrCh := s.startServers(httpListener, httpsListener)

	return s.waitForCompletion(ctx, httpServerErrCh, httpsServerErrCh)
}

func (s *Server) logServerStart() {
	s.logger.Infof("Starting HTTP server on %s", s.httpServer.Addr)

	if s.config.TLSEnabled {
		s.logger.Infof("Starting HTTPS server on %s", s.httpsServer.Addr)
	}
}

func (s *Server) startServers(
	httpListener, httpsListener net.Listener,
) (<-chan error, <-chan error) {
	httpServerErrCh := make(chan error, 1)
	httpsServerErrCh := make(chan error, 1)

	go func() {
		err := s.httpServer.Serve(httpListener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpServerErrCh <- ctxerrors.Wrap(err, "HTTP server error")
		}
	}()

	if s.config.TLSEnabled {
		go func() {
			err := s.httpsServer.ServeTLS(httpsListener, s.config.TLSCertFile, s.config.TLSKeyFile)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				httpsServerErrCh <- ctxerrors.Wrap(err, "HTTPS server error")
			}
		}()
	}

	return httpServerErrCh, httpsServerErrCh
}

func (s *Server) waitForCompletion(
	ctx context.Context,
	httpServerErrCh, httpsServerErrCh <-chan error,
) error {
	select {
	case <-ctx.Done():
		s.logger.Info("Context cancelled, shutting down servers")

		return ctxerrors.Wrap(ctx.Err(), "context cancelled")
	case <-s.doneCh:
		s.logger.Info("Stop called, shutting down servers")

		return nil
	case err := <-httpServerErrCh:
		return err
	case err := <-httpsServerErrCh:
		return err
	}
}

// Stop stops the HTTP server gracefully - can only be called once
//
//nolint:funlen // Graceful shutdown logic requires length
func (s *Server) Stop(ctx context.Context) error {
	var err error

	s.stopOnce.Do(func() {
		s.logger.Info("Starting graceful shutdown")

		// Close done channel to signal shutdown
		close(s.doneCh)

		// Create shutdown context with timeout from config
		shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
		defer cancel()

		// Collect all shutdown errors
		var (
			shutdownErrors []error
			wg             sync.WaitGroup
		)

		// Helper function to add errors safely
		addError := func(shutdownErr error) {
			if shutdownErr != nil {
				shutdownErrors = append(shutdownErrors, shutdownErr)
			}
		}

		// Shutdown HTTP server
		if s.httpServer != nil {
			wg.Add(1)

			go func() {
				defer wg.Done()

				s.logger.Info("Shutting down HTTP server")

				if shutdownErr := s.httpServer.Shutdown(shutdownCtx); shutdownErr != nil {
					s.logger.Errorf("HTTP server shutdown error: %v", shutdownErr)
					addError(ctxerrors.Wrap(shutdownErr, "failed to shutdown HTTP server"))
				} else {
					s.logger.Info("HTTP server shutdown completed")
				}
			}()
		}

		// Shutdown HTTPS server if TLS is enabled
		if s.config.TLSEnabled && s.httpsServer != nil {
			wg.Add(1)

			go func() {
				defer wg.Done()

				s.logger.Info("Shutting down HTTPS server")

				if shutdownErr := s.httpsServer.Shutdown(shutdownCtx); shutdownErr != nil {
					s.logger.Errorf("HTTPS server shutdown error: %v", shutdownErr)
					addError(ctxerrors.Wrap(shutdownErr, "failed to shutdown HTTPS server"))
				} else {
					s.logger.Info("HTTPS server shutdown completed")
				}
			}()
		}

		// Wait for all servers to shutdown
		wg.Wait()

		// Close listeners
		s.closeListeners(&shutdownErrors)

		// Return first error if any occurred
		if len(shutdownErrors) > 0 {
			err = shutdownErrors[0]

			const additionalErrorIndexOffset = 2
			for i, shutdownErr := range shutdownErrors[1:] {
				s.logger.Errorf("Additional shutdown error %d: %v", i+additionalErrorIndexOffset, shutdownErr)
			}
		}
	})

	return err
}

// closeListeners closes both HTTP and HTTPS listeners
func (s *Server) closeListeners(
	shutdownErrors *[]error,
) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	// Helper function to close a listener safely
	closeListener := func(listener net.Listener, name string) {
		if listener == nil {
			return
		}

		if closeErr := listener.Close(); closeErr != nil {
			// Check if it's just an "already closed" error
			if !strings.Contains(closeErr.Error(), "use of closed network connection") {
				s.logger.Errorf("Failed to close %s listener: %v", name, closeErr)
				*shutdownErrors = append(*shutdownErrors,
					ctxerrors.Wrapf(closeErr, "failed to close %s listener", name))
			} else {
				s.logger.Debugf("%s listener was already closed by server shutdown", name)
			}
		} else {
			s.logger.Infof("%s listener closed", name)
		}
	}

	// Close both listeners
	closeListener(s.httpListener, "HTTP")
	closeListener(s.httpsListener, "HTTPS")
}

func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

func (s *Server) GetLogger() *logrus.Logger {
	return s.logger
}

func (s *Server) GetRootGroup() *Group {
	s.rootGroupOnce.Do(func() {
		if s.router == nil {
			s.router = &Router{} // Initialize empty router if needed
		}

		if s.router.rootGroup == nil {
			s.router.rootGroup = NewGroup(
				s.mux,
				"/",
				s.logger,
			)
		}
	})

	return s.router.rootGroup
}

// GetListenerAddr returns the HTTP listener address if the server is running
// Deprecated: Use GetHTTPListenerAddr() instead
func (s *Server) GetListenerAddr() net.Addr {
	return s.GetHTTPListenerAddr()
}

// GetHTTPListenerAddr returns the HTTP listener address if the server is running
func (s *Server) GetHTTPListenerAddr() net.Addr {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()

	if s.httpListener == nil {
		return nil
	}

	return s.httpListener.Addr()
}

// GetHTTPSListenerAddr returns the HTTPS listener address if the server is running with TLS
func (s *Server) GetHTTPSListenerAddr() net.Addr {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()

	if s.httpsListener == nil {
		return nil
	}

	return s.httpsListener.Addr()
}
