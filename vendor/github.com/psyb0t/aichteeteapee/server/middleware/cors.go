package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/psyb0t/aichteeteapee"
)

// CORSConfig holds configuration for CORS middleware
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	MaxAge           int
	AllowCredentials bool
	AllowAllOrigins  bool
}

type CORSOption func(*CORSConfig)

// WithAllowedOrigins sets the allowed origins
func WithAllowedOrigins(origins ...string) CORSOption {
	return func(c *CORSConfig) {
		c.AllowedOrigins = origins
		c.AllowAllOrigins = false
	}
}

// WithAllowedMethods sets the allowed HTTP methods
func WithAllowedMethods(methods ...string) CORSOption {
	return func(c *CORSConfig) {
		c.AllowedMethods = methods
	}
}

// WithAllowedHeaders sets the allowed headers
func WithAllowedHeaders(headers ...string) CORSOption {
	return func(c *CORSConfig) {
		c.AllowedHeaders = headers
	}
}

// WithExposedHeaders sets the exposed headers
func WithExposedHeaders(headers ...string) CORSOption {
	return func(c *CORSConfig) {
		c.ExposedHeaders = headers
	}
}

// WithMaxAge sets the max age for preflight requests
func WithMaxAge(age int) CORSOption {
	return func(c *CORSConfig) {
		c.MaxAge = age
	}
}

// WithAllowCredentials enables credentials support
func WithAllowCredentials(allow bool) CORSOption {
	return func(c *CORSConfig) {
		c.AllowCredentials = allow
	}
}

// WithAllowAllOrigins allows all origins (sets Access-Control-Allow-Origin: *)
func WithAllowAllOrigins() CORSOption {
	return func(c *CORSConfig) {
		c.AllowAllOrigins = true
		c.AllowedOrigins = nil
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing with configurable options
//
//nolint:cyclop,funlen // Multiple if statements for header configuration is acceptable
func CORS(opts ...CORSOption) Middleware {
	config := &CORSConfig{
		AllowedOrigins:   []string{},
		AllowedMethods:   strings.Split(aichteeteapee.GetDefaultCORSAllowMethods(), ", "),
		AllowedHeaders:   strings.Split(aichteeteapee.GetDefaultCORSAllowHeaders(), ", "),
		ExposedHeaders:   []string{},
		MaxAge:           aichteeteapee.DefaultCORSMaxAge,
		AllowCredentials: false,
		AllowAllOrigins:  true, // Default behavior - allow all origins
	}

	for _, opt := range opts {
		opt(config)
	}

	// Create allowed origins map for efficient lookup
	allowedMap := make(map[string]bool)
	for _, origin := range config.AllowedOrigins {
		allowedMap[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get(aichteeteapee.HeaderNameOrigin)

			// Handle origin
			if config.AllowAllOrigins {
				w.Header().Set(
					aichteeteapee.HeaderNameAccessControlAllowOrigin,
					aichteeteapee.DefaultCORSAllowOriginAll,
				)
			} else if len(config.AllowedOrigins) > 0 && allowedMap[origin] {
				w.Header().Set(
					aichteeteapee.HeaderNameAccessControlAllowOrigin,
					origin,
				)
				w.Header().Set(
					aichteeteapee.HeaderNameVary,
					aichteeteapee.HeaderNameOrigin,
				)
			}

			if len(config.AllowedMethods) > 0 {
				w.Header().Set(
					aichteeteapee.HeaderNameAccessControlAllowMethods,
					strings.Join(config.AllowedMethods, ", "),
				)
			}

			if len(config.AllowedHeaders) > 0 {
				w.Header().Set(
					aichteeteapee.HeaderNameAccessControlAllowHeaders,
					strings.Join(config.AllowedHeaders, ", "),
				)
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set(
					aichteeteapee.HeaderNameAccessControlExposeHeaders,
					strings.Join(config.ExposedHeaders, ", "),
				)
			}

			w.Header().Set(
				aichteeteapee.HeaderNameAccessControlMaxAge,
				strconv.Itoa(config.MaxAge),
			)

			if config.AllowCredentials {
				w.Header().Set(
					aichteeteapee.HeaderNameAccessControlAllowCredentials,
					"true",
				)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
