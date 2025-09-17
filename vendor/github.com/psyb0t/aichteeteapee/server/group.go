package server

import (
	"net/http"
	"path"
	"strings"

	"github.com/psyb0t/aichteeteapee/server/middleware"
	"github.com/sirupsen/logrus"
)

type Group struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []middleware.Middleware
	logger      *logrus.Logger
}

func NewGroup(
	mux *http.ServeMux,
	prefix string,
	logger *logrus.Logger,
	middlewares ...middleware.Middleware,
) *Group {
	return &Group{
		mux:         mux,
		prefix:      cleanPrefix(prefix),
		middlewares: middlewares,
		logger:      logger,
	}
}

func (g *Group) Use(middlewares ...middleware.Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *Group) Group(
	subPrefix string,
	middlewares ...middleware.Middleware,
) *Group {
	// Combine parent middlewares with new ones
	combined := make([]middleware.Middleware, 0, len(g.middlewares)+len(middlewares))
	combined = append(combined, g.middlewares...)
	combined = append(combined, middlewares...)

	return &Group{
		mux:         g.mux,
		prefix:      joinPaths(g.prefix, subPrefix),
		middlewares: combined,
		logger:      g.logger,
	}
}

// Handle registers a handler for the given method and pattern
func (g *Group) Handle(
	method, pattern string,
	handler http.Handler,
	middlewares ...middleware.Middleware,
) {
	fullPath := joinPaths(g.prefix, pattern)
	route := buildRoute(method, fullPath)

	// Combine group middlewares with route-specific ones
	allMiddlewares := make([]middleware.Middleware, 0, len(g.middlewares)+len(middlewares))
	allMiddlewares = append(allMiddlewares, g.middlewares...)
	allMiddlewares = append(allMiddlewares, middlewares...)

	// Chain all middlewares
	finalHandler := middleware.Chain(handler, allMiddlewares...)

	g.logger.Debugf("Registering route: %s", route)
	g.mux.Handle(route, finalHandler)
}

// HandleFunc is a convenience method for registering http.HandlerFunc
func (g *Group) HandleFunc(
	method, pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.Handle(
		method,
		pattern,
		handler,
		middlewares...,
	)
}

func (g *Group) GET(
	pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.HandleFunc(
		http.MethodGet,
		pattern,
		handler,
		middlewares...,
	)
}

func (g *Group) POST(
	pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.HandleFunc(
		http.MethodPost,
		pattern,
		handler,
		middlewares...,
	)
}

func (g *Group) PUT(
	pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.HandleFunc(
		http.MethodPut,
		pattern,
		handler,
		middlewares...,
	)
}

func (g *Group) PATCH(
	pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.HandleFunc(
		http.MethodPatch,
		pattern,
		handler,
		middlewares...,
	)
}

func (g *Group) DELETE(
	pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.HandleFunc(
		http.MethodDelete,
		pattern,
		handler,
		middlewares...,
	)
}

func (g *Group) OPTIONS(
	pattern string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	g.HandleFunc(
		http.MethodOptions,
		pattern,
		handler,
		middlewares...,
	)
}

// Helper functions for path manipulation

// cleanPrefix normalizes a URL prefix by removing trailing slashes and cleaning the path.
// Returns empty string for root paths ("" or "/").
func cleanPrefix(prefix string) string {
	if prefix == "" || prefix == "/" {
		return ""
	}

	cleaned := path.Clean(prefix)
	if cleaned == "." {
		return ""
	}

	return strings.TrimSuffix(cleaned, "/")
}

// joinPaths combines a base path with a sub path, ensuring proper slash handling.
// Returns the base path if sub is empty or root.
func joinPaths(
	base, sub string,
) string {
	if sub == "" || sub == "/" {
		return base
	}

	// Ensure leading slash on sub path
	if !strings.HasPrefix(sub, "/") {
		sub = "/" + sub
	}

	// Combine and clean
	result := base + sub
	if result == "" {
		return "/"
	}

	return result
}

// buildRoute constructs a complete route pattern by combining HTTP method and path.
// Returns just the path if method is empty, ensuring path defaults to "/" if empty.
func buildRoute(
	method, path string,
) string {
	if path == "" {
		path = "/"
	}

	method = strings.TrimSpace(strings.ToUpper(method))
	if method == "" {
		return path
	}

	return method + " " + path
}
