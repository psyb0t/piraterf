# aichteeteapee

> **Pronounced "HTTP"** - because sometimes the best fucking code comes with wordplay.

**aichteeteapee** is a batteries-included HTTP utilities library that gets you from `go mod init` to working server with minimal configuration. Built on the philosophy of sane defaults, zero boilerplate, and easy customization.

Perfect for:
- 🚀 **Rapid prototyping** with solid foundations
- 🏗️ **Microservices** that need HTTP + WebSocket capabilities
- 📡 **APIs** requiring file uploads, static serving, and real-time features
- 🛠️ **Any Go project** that wants HTTP functionality without the fucking boilerplate

## Table of Contents
- [Quick Start - Zero to Hero](#quick-start---zero-to-hero)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Programmatic Config](#programmatic-config)
- [Key Features](#key-features)
- [Advanced Usage](#advanced-usage)
  - [Custom Middleware](#custom-middleware)
  - [WebSocket Events](#websocket-events)
  - [Unix Socket Bridge](#unix-socket-bridge---external-tool-integration)
  - [File Upload Processing](#file-upload-processing)
  - [Advanced Middleware Features](#advanced-middleware-features)
  - [Enhanced WebSocket Events](#enhanced-websocket-events)
- [Security Warnings ⚠️](#security-warnings-️)
- [License](#license)

## Quick Start - Zero to Hero

**Minimal example:**

```go
package main

import (
    "context"
    "net/http"
    "github.com/psyb0t/aichteeteapee/server"
)

func main() {
    srv, _ := server.New()
    router := server.Router{
        Groups: []server.GroupConfig{{
            Path: "/",
            Routes: []server.RouteConfig{
                {Method: "GET", Path: "/", Handler: func(w http.ResponseWriter, r *http.Request) {
                    w.Write([]byte("Hello, World!"))
                }},
            },
        }},
    }
    srv.Start(context.Background(), router)
}
```

**With features:**

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/psyb0t/aichteeteapee/server"
    "github.com/psyb0t/aichteeteapee/server/middleware"
    "github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
)

func main() {
    // Create server with sane defaults
    srv, err := server.New()
    if err != nil {
        log.Fatal("Failed to create server:", err)
    }

    // Create WebSocket hub for real-time features
    chatHub := wshub.NewHub("chat")

    // Build router with all the features
    router := server.Router{
        Middlewares: []middleware.Middleware{
            middleware.Logger(),          // Request logging
            middleware.Recovery(),        // Panic recovery
            middleware.CORS(),           // Smart CORS handling
            middleware.SecurityHeaders(), // Security headers
        },
        Groups: []server.GroupConfig{
            {
                Path: "/",
                Routes: []server.RouteConfig{
                    {Method: "GET", Path: "/", Handler: homeHandler},
                    {Method: "GET", Path: "/health", Handler: srv.HealthHandler()},
                    {Method: "POST", Path: "/upload", Handler: srv.FileUploadHandler("./uploads")},
                },
            },
            {
                Path: "/ws",
                Routes: []server.RouteConfig{
                    {Method: "GET", Path: "/chat", Handler: wshub.UpgradeHandler(chatHub)},
                },
            },
        },
        Static: []server.StaticRouteConfig{
            {
                URLPath:    "/static",
                LocalPath:  "./public",
                DirectoryIndexingType: server.DirectoryIndexingTypeHTML,
            },
        },
    }

    // Start server - HTTP on :8080, HTTPS on :8443, graceful shutdown
    log.Fatal(srv.Start(context.Background(), router))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Hello, World!", "status": "ok"}`))
}
```

That's it! You now have:
- ✅ HTTP + HTTPS servers running
- ✅ CORS, security headers, request logging
- ✅ File uploads with processing hooks & filename options
- ✅ Static file serving with directory browsing (HTML/JSON)
- ✅ WebSocket hub for real-time features with event metadata
- ✅ Request ID generation & extraction utilities
- ✅ Content-type enforcement middleware
- ✅ Timeout middleware with presets (short/default/long)
- ✅ Granular security header control
- ✅ Health checks and echo endpoints
- ✅ Unix socket bridge for external tool integration
- ✅ Graceful shutdown handling

## Configuration

### Environment Variables

```bash
# Server settings
HTTP_SERVER_LISTENADDRESS=127.0.0.1:8080
HTTP_SERVER_TLSLISTENADDRESS=127.0.0.1:8443
HTTP_SERVER_SERVICENAME=MyAPI

# Security
HTTP_SERVER_TLSENABLED=true
HTTP_SERVER_TLSCERTFILE=./certs/server.crt
HTTP_SERVER_TLSKEYFILE=./certs/server.key

# Timeouts (durations)
HTTP_SERVER_READTIMEOUT=30s
HTTP_SERVER_READHEADERTIMEOUT=10s
HTTP_SERVER_WRITETIMEOUT=30s
HTTP_SERVER_IDLETIMEOUT=60s
HTTP_SERVER_MAXHEADERBYTES=1048576
HTTP_SERVER_SHUTDOWNTIMEOUT=30s

# File uploads
HTTP_SERVER_FILEUPLOADMAXMEMORY=52428800  # 50MB in bytes
```

### Programmatic Config

```go
srv, err := server.NewWithConfig(server.Config{
    ListenAddress:       "0.0.0.0:8080",
    TLSListenAddress:    "0.0.0.0:8443",
    ServiceName:         "ProductionAPI",
    ReadTimeout:         30 * time.Second,
    ReadHeaderTimeout:   10 * time.Second,
    WriteTimeout:        30 * time.Second,
    IdleTimeout:         60 * time.Second,
    MaxHeaderBytes:      1 << 20, // 1MB
    ShutdownTimeout:     30 * time.Second,
    FileUploadMaxMemory: 100 << 20, // 100MB
    TLSEnabled:          true,
    TLSCertFile:         "/etc/ssl/certs/api.crt",
    TLSKeyFile:          "/etc/ssl/private/api.key",
})
```

## Key Features

### 🎯 **Zero-Config Defaults**
- **Secure defaults**: HTTPS, security headers, CORS, timeouts
- **Graceful shutdown**: Proper resource cleanup and connection draining
- **Structured logging**: Consistent field names with request tracing
- **Health checks**: Built-in `/health` and `/echo` endpoints

### 🌐 **Advanced HTTP Server**
- **Route groups**: Organize routes with shared middleware and configuration
- **Static files**: Automatic serving with configurable directory indexing
- **File uploads**: Built-in multipart handling with postprocessing hooks
- **Middleware system**: Composable, reusable middleware with proper ordering

### 🛡️ **Built-in Middleware**
- **CORS**: Cross-origin request handling for browser compatibility
- **Basic Auth**: Simple authentication with configurable realms
- **Security Headers**: HSTS, CSP, X-Frame-Options, and more
- **Logger**: Structured request logging with configurable fields
- **Recovery**: Panic recovery with stack traces and custom handlers
- **Request ID**: Automatic generation and extraction utilities
- **Timeout**: Configurable timeouts with presets (short/default/long)
- **Content-Type Enforcement**: API protection with configurable types

### ⚡ **WebSocket Systems**
> **Note**: WebSocket functionality lives in the `dabluveees` package - pronounced "dub-ell-vee-ess" (double-v-s = WS). More wordplay because memorable imports are better than boring ones.

**WebSocket Hub** (for real-time applications):
- **Multi-client management**: Automatic connection lifecycle handling
- **Event-driven architecture**: Type-safe event system with JSON marshaling
- **Broadcast capabilities**: Send to all clients, specific clients, or groups
- **Connection metadata**: Per-connection data storage and retrieval

**Unix Socket Bridge** (for external tool integration):
- **Bidirectional bridge**: WebSocket ↔ Unix domain sockets
- **External tool integration**: Shell scripts, CLI tools, other processes
- **File-based communication**: Simple read/write operations
- **Tool chaining**: Connect WebSocket apps to existing Unix toolchain

### 🛠️ **Developer Experience**
- **Sane defaults**: Works out of the box, customize when needed
- **File upload options**: UUID/DateTime/None filename prepending
- **Event metadata system**: Thread-safe WebSocket event enrichment
- **Error handling**: Proper HTTP status codes and JSON responses
- **Middleware composability**: Easy to chain and customize

## Advanced Usage

### Custom Middleware

```go
// Create your own middleware
func AuthMiddleware(secret string) middleware.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if !validateToken(token, secret) {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Use in route groups
{
    Path: "/api/protected",
    Middlewares: []middleware.Middleware{
        AuthMiddleware("your-secret-key"),
    },
    Routes: []server.RouteConfig{
        {Method: "GET", Path: "/profile", Handler: profileHandler},
    },
}
```

### WebSocket Events

```go
// Define your event types
const (
    EventTypeChatMessage dabluveees.EventType = "chat.message"
    EventTypeUserJoin    dabluveees.EventType = "user.join"
    EventTypeUserLeave   dabluveees.EventType = "user.leave"
)

// Create event handlers
chatHub := wshub.NewHub("chat")
chatHub.RegisterEventHandlers(map[dabluveees.EventType]wshub.EventHandler{
    EventTypeChatMessage: func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
        // Parse message data
        var messageData struct {
            Text   string `json:"text"`
            UserID string `json:"userId"`
        }
        if err := json.Unmarshal(event.Data, &messageData); err != nil {
            return err
        }

        // Add timestamp and broadcast
        event.Metadata.Set("timestamp", time.Now().Unix())
        hub.BroadcastToAll(event)
        return nil
    },
})
```

### Unix Socket Bridge - External Tool Integration

```go
// Unix socket bridge for external tool integration
import "github.com/psyb0t/aichteeteapee/server/dabluvee-es/wsunixbridge"

// Create Unix bridge handler
bridgeHandler := wsunixbridge.NewUpgradeHandler(
    "./sockets",  // Directory for Unix socket files
    func(connection *wsunixbridge.Connection) error {
        log.Printf("Unix bridge connection established: %s", connection.ID)
        log.Printf("Writer socket: %s", connection.WriterUnixSock.Path)
        log.Printf("Reader socket: %s", connection.ReaderUnixSock.Path)
        return nil
    },
)

// Add to routes
{Method: "GET", Path: "/unixsock", Handler: bridgeHandler}

// External tools can now:
// 1. Connect to writer socket to receive WebSocket data
// 2. Write to reader socket to send data to WebSocket
//
// Example: echo "Hello from external tool" | socat - UNIX-CONNECT:./sockets/connection-id_input
```

### File Upload Processing

```go
// Advanced file upload configuration
uploadHandler := srv.FileUploadHandler("./uploads",
    // Custom postprocessor for file processing
    server.WithFileUploadHandlerPostprocessor(func(
        response map[string]any,
        request *http.Request,
    ) (map[string]any, error) {
        // Add custom metadata to upload response
        response["processed_at"] = time.Now().Unix()
        response["user_ip"] = request.RemoteAddr
        return response, nil
    }),

    // Filename prepending options
    server.WithFilenamePrependType(server.FilenamePrependTypeDateTime), // Y_M_D_H_I_S_
    // Alternative: server.FilenamePrependTypeUUID (default)
    // Alternative: server.FilenamePrependTypeNone
)
```

### Advanced Middleware Features

```go
import (
    "github.com/psyb0t/aichteeteapee/server/middleware"
    "github.com/sirupsen/logrus"
)

// Request ID middleware with utility functions
router := server.Router{
    Middlewares: []middleware.Middleware{
        middleware.RequestID(), // Automatic request ID generation
    },
    Groups: []server.GroupConfig{
        {
            Path: "/api",
            Routes: []server.RouteConfig{
                {
                    Method: "GET",
                    Path: "/status",
                    Handler: func(w http.ResponseWriter, r *http.Request) {
                        // Extract request ID from context
                        reqID := middleware.GetRequestID(r)

                        response := map[string]string{
                            "status":     "ok",
                            "request_id": reqID,
                        }
                        json.NewEncoder(w).Encode(response)
                    },
                },
            },
        },
    },
}

// Content-Type enforcement for APIs
apiGroup := server.GroupConfig{
    Path: "/api",
    Middlewares: []middleware.Middleware{
        // Only allow JSON requests
        middleware.EnforceRequestContentType("application/json"),
        // Alternative: allow multiple types
        // middleware.EnforceRequestContentType("application/json", "application/xml"),
        // Shortcut for JSON-only APIs
        // middleware.EnforceRequestContentTypeJSON(),
    },
    Routes: []server.RouteConfig{
        {Method: "POST", Path: "/users", Handler: createUserHandler},
    },
}

// Timeout middleware with presets
timeoutGroup := server.GroupConfig{
    Path: "/slow-api",
    Middlewares: []middleware.Middleware{
        middleware.Timeout(
            middleware.WithLongTimeout(), // 5 minutes
            // Alternative presets:
            // middleware.WithShortTimeout(),   // 5 seconds
            // middleware.WithDefaultTimeout(), // 30 seconds
            // middleware.WithTimeout(2 * time.Minute), // Custom
        ),
    },
    Routes: []server.RouteConfig{
        {Method: "POST", Path: "/batch-process", Handler: batchProcessHandler},
    },
}

// Security headers with granular control
secureGroup := server.GroupConfig{
    Path: "/secure",
    Middlewares: []middleware.Middleware{
        middleware.SecurityHeaders(
            // Enable specific headers
            middleware.WithXContentTypeOptions("nosniff"),
            middleware.WithXFrameOptions("DENY"),
            middleware.WithStrictTransportSecurity("max-age=31536000; includeSubDomains"),

            // Disable unwanted headers
            middleware.DisableXXSSProtection(), // If you handle XSS at app level
            middleware.DisableCSP(),            // If you have custom CSP
        ),
    },
}
```

### Enhanced WebSocket Events

```go
import "github.com/psyb0t/aichteeteapee/server/dabluvee-es"

// Event creation with metadata and utility methods
hub := wshub.NewHub("notifications")

hub.RegisterEventHandler(EventTypeNewMessage, func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
    // Add metadata to events
    enrichedEvent := event.
        WithMetadata("server_id", "api-01").
        WithMetadata("processing_time", time.Now().Unix()).
        WithTimestamp(time.Now().Unix()) // Override timestamp

    // Check if event is recent (within last 60 seconds)
    if enrichedEvent.IsRecent(60) {
        // Get event time as Go time.Time
        eventTime := enrichedEvent.GetTime()
        log.Printf("Processing recent event from %v", eventTime)

        // Broadcast with enriched metadata
        hub.BroadcastToAll(&enrichedEvent)
    }

    return nil
})

// Built-in event types available
const (
    // System events
    dabluveees.EventTypeSystemLog   // "system.log"
    dabluveees.EventTypeShellExec   // "shell.exec"
    dabluveees.EventTypeEchoRequest // "echo.request"
    dabluveees.EventTypeEchoReply   // "echo.reply"
    dabluveees.EventTypeError       // "error"
)
```

## Security Warnings ⚠️

**READ THIS SHIT CAREFULLY** - Security is not a fucking joke:

### 🔥 **CRITICAL - Authentication & Authorization**
- **NEVER** run without authentication in production
- **ALWAYS** validate user permissions before accessing resources
- **USE** HTTPS in production - HTTP is not secure
- **IMPLEMENT** proper session management and token validation

### 🔒 **File Upload Security**
- **VALIDATE** file types and sizes - don't trust client data
- **SCAN** uploads for malware before processing
- **STORE** uploads outside web root to prevent direct access
- **LIMIT** file extensions and use whitelist, not blacklist

### 🛡️ **Path Traversal Protection**
- Library includes protection, but **ALWAYS** validate custom file paths
- **NEVER** trust user input for file system operations
- **USE** absolute paths and proper validation for file access

### 🚨 **WebSocket Security**
- **AUTHENTICATE** WebSocket connections - they bypass normal HTTP auth
- **VALIDATE** all event data - treat it as untrusted input
- **IMPLEMENT** rate limiting for WebSocket messages
- **MONITOR** connection counts to prevent DoS attacks

### 🔐 **Production Checklist**
- [ ] HTTPS configured with valid certificates
- [ ] Security headers properly configured
- [ ] Authentication middleware on protected routes
- [ ] File upload validation and scanning
- [ ] Proper error handling (don't leak internal info)
- [ ] Request logging and monitoring in place
- [ ] Rate limiting configured

## License

MIT License. See LICENSE file for details.