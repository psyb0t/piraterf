package aichteeteapee

import (
	"net/http"
	"strings"
	"time"
)

const (
	// Server defaults
	DefaultHTTPServerListenAddress     = "127.0.0.1:8080"
	DefaultHTTPServerReadTimeout       = 15 * time.Second
	DefaultHTTPServerReadHeaderTimeout = 10 * time.Second
	DefaultHTTPServerWriteTimeout      = 30 * time.Second
	DefaultHTTPServerIdleTimeout       = 60 * time.Second
	DefaultHTTPServerMaxHeaderBytes    = 1 << 20 // 1MB
	DefaultHTTPServerShutdownTimeout   = 10 * time.Second
	DefaultHTTPServerServiceName       = "http-server"

	// TLS Server defaults
	DefaultHTTPServerTLSEnabled       = false
	DefaultHTTPServerTLSListenAddress = "127.0.0.1:8443"
	DefaultHTTPServerTLSCertFile      = ""
	DefaultHTTPServerTLSKeyFile       = ""

	// Request defaults
	DefaultHTTPRequestTimeout = 30 * time.Second
	DefaultHTTPClientTimeout  = 30 * time.Second

	// CORS defaults
	DefaultCORSAllowOriginAll = "*"
	DefaultCORSMaxAge         = 86400 // 24 hours in seconds

	// Security header default values
	DefaultSecurityXContentTypeOptionsNoSniff = "nosniff"
	DefaultSecurityXFrameOptionsDeny          = "DENY"
	DefaultSecurityXXSSProtectionBlock        = "1; mode=block"
	DefaultSecurityStrictTransportSecurity    = "max-age=31536000; includeSubDomains"
	DefaultSecurityReferrerPolicyStrictOrigin = "strict-origin-when-cross-origin"

	// Authentication default values
	DefaultBasicRealmName      = "restricted"
	DefaultUnauthorizedMessage = "Unauthorized"

	// File upload defaults
	DefaultFileUploadMaxMemory = int64(32 << 20) // 32MB

	// WebSocket Client Configuration Defaults
	DefaultWebSocketClientSendBufferSize  = 256
	DefaultWebSocketClientReadBufferSize  = 1024
	DefaultWebSocketClientWriteBufferSize = 1024
	DefaultWebSocketClientReadLimit       = 1024 * 1024 // 1MB
	DefaultWebSocketClientReadTimeout     = 60 * time.Second
	DefaultWebSocketClientWriteTimeout    = 10 * time.Second
	DefaultWebSocketClientPingInterval    = 54 * time.Second
	DefaultWebSocketClientPongTimeout     = 60 * time.Second

	// WebSocket Handler Configuration Defaults
	DefaultWebSocketHandlerReadBufferSize    = 1024
	DefaultWebSocketHandlerWriteBufferSize   = 1024
	DefaultWebSocketHandlerHandshakeTimeout  = 45 * time.Second
	DefaultWebSocketHandlerEnableCompression = false
)

func GetDefaultCORSAllowMethods() string {
	return strings.Join([]string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}, ", ")
}

func GetDefaultCORSAllowHeaders() string {
	return strings.Join([]string{
		HeaderNameAuthorization,
		HeaderNameContentType,
		HeaderNameXRequestID,
	}, ", ")
}

// GetDefaultWebSocketCheckOrigin is the default origin checker for WebSocket connections
// WARNING: This allows all origins - configure for your security needs in production
func GetDefaultWebSocketCheckOrigin(_ *http.Request) bool {
	return true
}
