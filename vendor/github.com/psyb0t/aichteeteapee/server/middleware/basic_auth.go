package middleware

import (
	"context"
	"crypto/subtle"
	"maps"
	"net/http"

	"github.com/psyb0t/aichteeteapee"
)

// BasicAuthConfig holds configuration for basic auth middleware.
type BasicAuthConfig struct {
	Realm           string
	Users           map[string]string // username -> password mapping for users
	UnauthorizedMsg string
	Validator       func(username, password string) bool
	SkipPaths       map[string]bool
	UseConstantTime bool // Use constant-time comparison to prevent timing attacks
	SendChallenge   bool
	// Whether to send WWW-Authenticate header (triggers browser popup)
}

type BasicAuthOption func(*BasicAuthConfig)

// WithBasicAuthUsers sets username/password pairs.
func WithBasicAuthUsers(users map[string]string) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		if c.Users == nil {
			c.Users = make(map[string]string)
		}

		maps.Copy(c.Users, users)
	}
}

// WithBasicAuthRealm sets the authentication realm.
func WithBasicAuthRealm(realm string) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		c.Realm = realm
	}
}

// WithBasicAuthUnauthorizedMessage sets the unauthorized response message.
func WithBasicAuthUnauthorizedMessage(
	message string,
) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		c.UnauthorizedMsg = message
	}
}

// WithBasicAuthValidator sets a custom validation function.
func WithBasicAuthValidator(
	validator func(username, password string) bool,
) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		c.Validator = validator
	}
}

// WithBasicAuthSkipPaths sets paths to skip authentication.
func WithBasicAuthSkipPaths(paths ...string) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		if c.SkipPaths == nil {
			c.SkipPaths = make(map[string]bool)
		}

		for _, path := range paths {
			c.SkipPaths[path] = true
		}
	}
}

// WithConstantTimeComparison enables constant-time string comparison
// to prevent timing attacks.
func WithConstantTimeComparison(enable bool) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		c.UseConstantTime = enable
	}
}

// WithBasicAuthChallenge controls whether to send WWW-Authenticate
// header (browser popup).
func WithBasicAuthChallenge(sendChallenge bool) BasicAuthOption {
	return func(c *BasicAuthConfig) {
		c.SendChallenge = sendChallenge
	}
}

// BasicAuthMiddleware provides HTTP Basic Authentication
// with configurable options
//

func BasicAuth(opts ...BasicAuthOption) Middleware {
	config := &BasicAuthConfig{
		Realm:           aichteeteapee.DefaultBasicRealmName,
		Users:           make(map[string]string),
		UnauthorizedMsg: aichteeteapee.DefaultUnauthorizedMessage,
		Validator:       nil,
		SkipPaths:       make(map[string]bool),
		UseConstantTime: true, // Default to secure comparison
		SendChallenge:   true, // Default to sending challenge (browser popup)
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for specified paths
			if config.SkipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)

				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok {
				unauthorized(w, config)

				return
			}

			if !authenticateUser(config, user, pass) {
				unauthorized(w, config)

				return
			}

			// Authentication successful - add user to context
			ctx := context.WithValue(r.Context(), aichteeteapee.ContextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authenticateUser performs user authentication with the given config.
func authenticateUser(config *BasicAuthConfig, user, pass string) bool {
	// Use custom validator if provided
	if config.Validator != nil {
		return config.Validator(user, pass)
	}

	if len(config.Users) == 0 {
		return false
	}

	return authenticateWithUsers(config, user, pass)
}

// authenticateWithUsers performs authentication against the user map.
func authenticateWithUsers(config *BasicAuthConfig, user, pass string) bool {
	if config.UseConstantTime {
		return constantTimeAuth(config.Users, user, pass)
	}

	expectedPassword, exists := config.Users[user]

	return exists && pass == expectedPassword
}

// constantTimeAuth performs constant-time authentication
// to prevent timing attacks.
func constantTimeAuth(users map[string]string, user, pass string) bool {
	expectedPassword, exists := users[user]
	if !exists {
		// Use dummy password to ensure constant-time operation
		expectedPassword = "dummy-password-to-prevent-timing-attack"
	}

	passwordMatch := subtle.ConstantTimeCompare(
		[]byte(pass),
		[]byte(expectedPassword),
	) == 1

	return exists && passwordMatch
}

// unauthorized sends an unauthorized response.
func unauthorized(w http.ResponseWriter, config *BasicAuthConfig) {
	if config.SendChallenge {
		w.Header().Set(
			aichteeteapee.HeaderNameWWWAuthenticate,
			"Basic realm=\""+config.Realm+"\"",
		)
	}

	http.Error(
		w,
		config.UnauthorizedMsg,
		http.StatusUnauthorized,
	)
}
