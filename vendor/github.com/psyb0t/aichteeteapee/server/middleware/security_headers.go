package middleware

import (
	"net/http"

	"github.com/psyb0t/aichteeteapee"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	XContentTypeOptions        string
	XFrameOptions              string
	XXSSProtection             string
	StrictTransportSecurity    string
	ReferrerPolicy             string
	ContentSecurityPolicy      string
	DisableXContentTypeOptions bool
	DisableXFrameOptions       bool
	DisableXXSSProtection      bool
	DisableHSTS                bool
	DisableReferrerPolicy      bool
	DisableCSP                 bool
}

type SecurityHeadersOption func(*SecurityHeadersConfig)

func WithXContentTypeOptions(value string) SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.XContentTypeOptions = value
	}
}

func WithXFrameOptions(value string) SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.XFrameOptions = value
	}
}

func WithXXSSProtection(value string) SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.XXSSProtection = value
	}
}

func WithStrictTransportSecurity(value string) SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.StrictTransportSecurity = value
	}
}

// WithReferrerPolicy sets the Referrer-Policy header value
func WithReferrerPolicy(value string) SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.ReferrerPolicy = value
	}
}

// WithContentSecurityPolicy sets the Content-Security-Policy header value
func WithContentSecurityPolicy(value string) SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.ContentSecurityPolicy = value
	}
}

// DisableXContentTypeOptions disables the X-Content-Type-Options header
func DisableXContentTypeOptions() SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.DisableXContentTypeOptions = true
	}
}

// DisableXFrameOptions disables the X-Frame-Options header
func DisableXFrameOptions() SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.DisableXFrameOptions = true
	}
}

// DisableXXSSProtection disables the X-XSS-Protection header
func DisableXXSSProtection() SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.DisableXXSSProtection = true
	}
}

// DisableHSTS disables the Strict-Transport-Security header
func DisableHSTS() SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.DisableHSTS = true
	}
}

// DisableReferrerPolicy disables the Referrer-Policy header
func DisableReferrerPolicy() SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.DisableReferrerPolicy = true
	}
}

// DisableCSP disables the Content-Security-Policy header
func DisableCSP() SecurityHeadersOption {
	return func(c *SecurityHeadersConfig) {
		c.DisableCSP = true
	}
}

// SecurityHeadersMiddleware adds common security headers with default values
func SecurityHeaders(opts ...SecurityHeadersOption) Middleware {
	config := &SecurityHeadersConfig{
		XContentTypeOptions:     aichteeteapee.DefaultSecurityXContentTypeOptionsNoSniff,
		XFrameOptions:           aichteeteapee.DefaultSecurityXFrameOptionsDeny,
		XXSSProtection:          aichteeteapee.DefaultSecurityXXSSProtectionBlock,
		StrictTransportSecurity: aichteeteapee.DefaultSecurityStrictTransportSecurity,
		ReferrerPolicy:          aichteeteapee.DefaultSecurityReferrerPolicyStrictOrigin,
		ContentSecurityPolicy:   "",
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.DisableXContentTypeOptions {
				w.Header().Set(aichteeteapee.HeaderNameXContentTypeOptions, config.XContentTypeOptions)
			}

			if !config.DisableXFrameOptions {
				w.Header().Set(aichteeteapee.HeaderNameXFrameOptions, config.XFrameOptions)
			}

			if !config.DisableXXSSProtection {
				w.Header().Set(aichteeteapee.HeaderNameXXSSProtection, config.XXSSProtection)
			}

			if !config.DisableHSTS {
				w.Header().Set(aichteeteapee.HeaderNameStrictTransportSecurity, config.StrictTransportSecurity)
			}

			if !config.DisableReferrerPolicy {
				w.Header().Set(aichteeteapee.HeaderNameReferrerPolicy, config.ReferrerPolicy)
			}

			if !config.DisableCSP && config.ContentSecurityPolicy != "" {
				w.Header().Set(aichteeteapee.HeaderNameContentSecurityPolicy, config.ContentSecurityPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}
