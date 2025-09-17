package server

import (
	"net/http"

	"github.com/psyb0t/aichteeteapee/server/middleware"
)

type StaticRouteConfig struct {
	Dir                   string                // Directory to serve files from
	Path                  string                // URL path prefix to serve on
	DirectoryIndexingType DirectoryIndexingType // Directory indexing type: none, html, or json (default: none)
}

type Router struct {
	// Global middleware settings
	GlobalMiddlewares []middleware.Middleware

	// Multiple static file routes
	Static []StaticRouteConfig

	// Route groups
	Groups []GroupConfig

	// Internal field for root group (not serialized)
	rootGroup *Group
}

type GroupConfig struct {
	Path        string
	Middlewares []middleware.Middleware
	Routes      []RouteConfig
	Groups      []GroupConfig // Nested groups
}

type RouteConfig struct {
	Method  string           // http.MethodGet, http.MethodPost, etc.
	Path    string           // "/users/{id}", etc.
	Handler http.HandlerFunc // Handler function
}
