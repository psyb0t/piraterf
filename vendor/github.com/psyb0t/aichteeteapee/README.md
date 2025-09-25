# aichteeteapee 🌶️

_pronounced "HTTP" because comedic genius was involved here_

**📚 [API Reference](https://pkg.go.dev/github.com/psyb0t/aichteeteapee)**

## Table of Contents

- [🚀 30-Second Quick Start](#30-second-quick-start)
- [📦 Root Utilities](#the-root-utilities-use-anywhere)
- [🖥️ Full Server](#the-full-server-beast-mode)
- [🔌 WebSocket System](#websocket-system)
  - [WebSocket Hub (wshub)](#websocket-hub-wshub)
  - [Unix Socket Bridge (wsunixbridge)](#unix-socket-bridge-wsunixbridge)
- [📁 Static Files & Uploads](#static-files--uploads)
- [🛠️ Middleware System](#middleware-system)
- [⚙️ Configuration](#configuration)
- [🚨 Troubleshooting](#troubleshooting)
- [🚀 Production Deployment](#production-deployment)

## dafuq is dis bish?

**aichteeteapee** is a collection of HTTP utilities that don't suck. It's got two main parts:

1. **Root package**: Common HTTP utilities you can use anywhere - JSON responses, request parsing, headers, error codes, etc.
2. **Server package**: A complete batteries-included web server with WebSocket support, middleware, static files, and all that jazz

Use just the utilities with your existing server, or go full beast mode with the complete server. Your call.

## Installation

```bash
go get github.com/psyb0t/aichteeteapee
```

## 30-Second Quick Start

```bash
go mod init myapp && go get github.com/psyb0t/aichteeteapee
```

```go
package main

import (
    "context"
    "net/http"
    "github.com/psyb0t/aichteeteapee/server"
)

func main() {
    s, _ := server.New()

    router := &server.Router{
        Groups: []server.GroupConfig{{
            Path: "/",
            Routes: []server.RouteConfig{{
                Method: http.MethodGet,
                Path: "/",
                Handler: func(w http.ResponseWriter, r *http.Request) {
                    w.Write([]byte("Hello World!"))
                },
            }},
        }},
    }

    s.Start(context.Background(), router) // Server running on :8080
}
```

**BOOM!** You have a production-ready HTTP server with CORS, logging, security headers, and graceful shutdown.

## Core Types Reference

Quick reference for main types - see [full API docs](https://pkg.go.dev/github.com/psyb0t/aichteeteapee) for complete details.

<details>
<summary><strong>Server Package Types</strong></summary>

```go
// Server is the main HTTP server
type Server struct {
    // Exported methods:
    Start(ctx context.Context, router *Router) error
    Stop(ctx context.Context) error
    GetRootGroup() *Group
    GetMux() *http.ServeMux
    GetHTTPListenerAddr() net.Addr
    GetHTTPSListenerAddr() net.Addr

    // Built-in handlers:
    HealthHandler(w http.ResponseWriter, r *http.Request)
    EchoHandler(w http.ResponseWriter, r *http.Request)
    FileUploadHandler(uploadsDir string, opts ...FileUploadHandlerOption) http.HandlerFunc
}

// Router defines your complete server configuration
type Router struct {
    // Applied to all routes
    GlobalMiddlewares []middleware.Middleware
    // Static file serving configs
    Static           []StaticRouteConfig
    // Route groups
    Groups           []GroupConfig
}

// StaticRouteConfig for serving static files
type StaticRouteConfig struct {
    // "./static" - directory to serve
    Dir                   string
    // "/static" - URL path prefix
    Path                  string
    // HTML, JSON, or None
    DirectoryIndexingType DirectoryIndexingType
}

// GroupConfig for organizing routes
type GroupConfig struct {
    // "/api/v1" - group path prefix
    Path        string
    // Group-specific middleware
    Middlewares []middleware.Middleware
    // Routes in this group
    Routes      []RouteConfig
    // Nested groups (recursive)
    Groups      []GroupConfig
}

// RouteConfig defines individual routes
type RouteConfig struct {
    // http.MethodGet, http.MethodPost, etc.
    Method  string
    // "/users/{id}" - route pattern
    Path    string
    // Your handler function
    Handler http.HandlerFunc
}

// Group provides fluent API for route registration
type Group struct {
    // Methods:
    Use(middlewares ...middleware.Middleware)
    Group(subPrefix string, middlewares ...middleware.Middleware) *Group
    Handle(method, pattern string, handler http.Handler, middlewares ...middleware.Middleware)
    HandleFunc(method, pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
    GET(pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
    POST(pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
    PUT(pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
    PATCH(pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
    DELETE(pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
    OPTIONS(pattern string, handler http.HandlerFunc, middlewares ...middleware.Middleware)
}

// Server constructors
func New() (*Server, error)
func NewWithLogger(logger *logrus.Logger) (*Server, error)
func NewWithConfig(config Config) (*Server, error)
func NewWithConfigAndLogger(config Config, logger *logrus.Logger) (*Server, error)
```

</details>


<details>
<summary><strong>Middleware Package Types</strong></summary>

```go
// Middleware is just the standard http middleware pattern
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares around a handler
func Chain(h http.Handler, middlewares ...Middleware) http.Handler

// Built-in middlewares (all return Middleware)
func Recovery() Middleware                    // Panic recovery
func RequestID() Middleware                   // Request ID generation
func Logger(opts ...LoggerOption) Middleware // Request logging
func SecurityHeaders() Middleware             // Security headers (XSS, CSRF, etc.)
func CORS(opts ...CORSOption) Middleware     // CORS handling
func Timeout(duration time.Duration) Middleware // Request timeout
func BasicAuth(users map[string]string, opts ...BasicAuthOption) Middleware // Basic auth
func EnforceRequestContentType(contentType string) Middleware // Content-Type enforcement
```

</details>

## The Root Utilities (Use Anywhere)

The base `aichteeteapee` package gives you all the HTTP essentials:

```go
import "github.com/psyb0t/aichteeteapee"

// Pretty JSON responses with proper headers
aichteeteapee.WriteJSON(w, 200, map[string]string{"status": "winning"})

// Smart client IP extraction (handles proxies, load balancers, etc.)
clientIP := aichteeteapee.GetClientIP(r)

// Content type checking that actually works
if aichteeteapee.IsRequestContentTypeJSON(r) {
    // Handle JSON like a boss
}

// Request ID for tracing (if you set it in context)
requestID := aichteeteapee.GetRequestID(r)

// Predefined error responses that don't make you cry
aichteeteapee.WriteJSON(w, 404, aichteeteapee.ErrorResponseFileNotFound)
```

**What you get:**

- ✅ `WriteJSON()` - JSON responses with pretty formatting
- ✅ `GetClientIP()` - Smart IP extraction (X-Forwarded-For → X-Real-IP → RemoteAddr)
- ✅ `IsRequestContentTypeJSON/XML/FormData()` - Content type checking that works
- ✅ `GetRequestID()` - Request ID extraction from context
- ✅ HTTP header constants (`HeaderNameContentType`, etc.)
- ✅ Content type constants (`ContentTypeJSON`, etc.)
- ✅ Predefined error responses (`ErrorResponseBadRequest`, etc.)
- ✅ Context keys for request metadata

## The Full Server (Beast Mode)

Want everything? Here's a complete server setup:

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "github.com/psyb0t/aichteeteapee/server"
    "github.com/psyb0t/aichteeteapee/server/middleware"
    dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
    "github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
)

func main() {
    // Create server
    s, err := server.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create WebSocket hub for real-time features
    hub := wshub.NewHub("my-app")

    // Setup WebSocket event handlers
    hub.RegisterEventHandler(dabluveees.EventTypeEchoRequest, func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
        // Echo it back to the sender
        return client.SendEvent(dabluveees.NewEvent(dabluveees.EventTypeEchoReply, event.Data))
    })

    hub.RegisterEventHandler("file.delete", func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
        type deleteMsg struct {
            FilePath string `json:"filePath"`
        }

        var msg deleteMsg
        json.Unmarshal(event.Data, &msg)

        // Do the file delete
        if err := os.Remove(msg.FilePath); err != nil {
            // Broadcast error to all clients
            hub.BroadcastToAll(dabluveees.NewEvent("file.delete.error", map[string]string{
                "error": err.Error(),
                "file":  msg.FilePath,
            }))
            return nil
        }

        // Broadcast success to all clients
        hub.BroadcastToAll(dabluveees.NewEvent("file.delete.success", map[string]string{
            "file": msg.FilePath,
        }))
        return nil
    })

    // Define your complete server structure using Router struct
    router := &server.Router{
        GlobalMiddlewares: []middleware.Middleware{
            middleware.Recovery(),      // Panic recovery
            middleware.RequestID(),     // Request tracing
            middleware.Logger(),        // Request logging
            middleware.SecurityHeaders(), // Security headers
            middleware.CORS(),          // CORS handling
        },
        Static: []server.StaticRouteConfig{
            {
                Dir:  "./static",      // Serve static files
                Path: "/static",
            },
            {
                Dir:                   "./uploads",
                Path:                  "/files",
                DirectoryIndexingType: server.DirectoryIndexingTypeJSON, // Browseable uploads
            },
        },
        Groups: []server.GroupConfig{
            {
                Path: "/",
                Routes: []server.RouteConfig{
                    {
                        Method:  http.MethodGet,
                        Path:    "/",
                        Handler: func(w http.ResponseWriter, r *http.Request) {
                            w.Write([]byte("Welcome to the fucking show!"))
                        },
                    },
                    {
                        Method:  http.MethodGet,
                        Path:    "/ws",
                        Handler: dabluveees.UpgradeHandler(hub), // WebSocket endpoint
                    },
                    {
                        Method:  http.MethodPost,
                        Path:    "/upload",
                        Handler: s.FileUploadHandler("./uploads"), // File uploads
                    },
                },
            },
        },
    }

    // Start the beast
    log.Println("Starting server...")
    if err := s.Start(context.Background(), router); err != nil {
        log.Fatal(err)
    }
}
```

That's it. No, seriously. **THAT'S FUCKING IT.**

You now have:

- ✅ HTTP server on `:8080`
- ✅ HTTPS server on `:8443` (if you enable and configure TLS certs)
- ✅ CORS that doesn't hate you
- ✅ Request logging that makes sense
- ✅ Security headers that actually secure
- ✅ Static file serving from `./static`
- ✅ File uploads at `/upload` with UUID filenames
- ✅ Directory browsing at `/files` (JSON format)
- ✅ WebSocket support at `/ws`
- ✅ Panic recovery (your server won't die)
- ✅ Request ID tracing
- ✅ Graceful shutdown

## But I Want Simple Shit

Fine, you minimalist bastard:

```go
func main() {
    s, _ := server.New()

    // Just use the Router struct with basic routes
    router := &server.Router{
        Groups: []server.GroupConfig{
            {
                Path: "/",
                Routes: []server.RouteConfig{
                    {
                        Method: http.MethodGet,
                        Path:   "/",
                        Handler: func(w http.ResponseWriter, r *http.Request) {
                            aichteeteapee.WriteJSON(w, 200, map[string]string{
                                "message": "Hello World",
                            })
                        },
                    },
                },
            },
        },
    }

    s.Start(context.Background(), router)
}
```

## WebSocket System

The WebSocket system in aichteeteapee is organized into the **dabluvee-es** package (_pronounced "WS" like double-v-S, because why the fuck not_):

- **`server/dabluvee-es`** - Base WebSocket configuration, events, and utilities
- **`server/dabluvee-es/wshub`** - Event-driven WebSocket hub system
- **`server/dabluvee-es/wsunixbridge`** - WebSocket to Unix socket bridge

### WebSocket Hub (wshub)

The hub system provides event-driven WebSocket communication with client management and broadcasting.

**Basic Hub Setup:**

```go
import (
    dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
    "github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
)

// Create a hub
hub := wshub.NewHub("my-app")

// Register event handlers
hub.RegisterEventHandler(dabluveees.EventTypeEchoRequest, func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
    // Echo back to sender
    return client.SendEvent(dabluveees.NewEvent(dabluveees.EventTypeEchoReply, event.Data))
})

// Custom event handlers
hub.RegisterEventHandler("user.login", func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
    // Broadcast to all clients
    return hub.BroadcastToAll(dabluveees.NewEvent("user.online", map[string]string{
        "userId": "123",
        "status": "online",
    }))
})

// Add to your router
router := &server.Router{
    Groups: []server.GroupConfig{
        {
            Path: "/",
            Routes: []server.RouteConfig{
                {
                    Method:  http.MethodGet,
                    Path:    "/ws",
                    Handler: dabluveees.UpgradeHandler(hub),
                },
            },
        },
    },
}
```

**Event Creation:**

```go
// Create events with any data
event := dabluveees.NewEvent("user.message", map[string]string{
    "message": "Hello world!",
    "username": "john",
})

// Events have metadata support
event.Metadata.Set("priority", "high")
event.Metadata.Set("source", "web-client")
```

**Broadcasting Options:**

- `client.SendEvent(event)` - Send to specific client only
- `hub.BroadcastToAll(event)` - Send to everyone in the hub
- `hub.BroadcastToClients(clientIDs, event)` - Send to specific clients

**Multiple Hubs:**

```go
chatHub := wshub.NewHub("chat")
notificationHub := wshub.NewHub("notifications")

// Different endpoints for different purposes
router := &server.Router{
    Groups: []server.GroupConfig{
        {
            Path: "/",
            Routes: []server.RouteConfig{
                {Method: http.MethodGet, Path: "/ws/chat", Handler: dabluveees.UpgradeHandler(chatHub)},
                {Method: http.MethodGet, Path: "/ws/notifications", Handler: dabluveees.UpgradeHandler(notificationHub)},
            },
        },
    },
}
```

### Unix Socket Bridge (wsunixbridge)

The Unix socket bridge creates Unix domain sockets that external tools can connect to for bidirectional communication with WebSocket clients.

**How it works:**

- **WriterUnixSock (`_output`)**: WebSocket data → Unix socket → external tools READ
- **ReaderUnixSock (`_input`)**: External tools WRITE → Unix socket → WebSocket

**Basic Setup:**

```go
import "github.com/psyb0t/aichteeteapee/server/dabluvee-es/wsunixbridge"

socketsDir := "./sockets"

// Connection handler (optional)
connectionHandler := func(conn *wsunixbridge.Connection) error {
    log.Printf("Unix socket bridge connection: %s", conn.ID)
    log.Printf("Output socket: %s", conn.WriterUnixSock.Path)
    log.Printf("Input socket: %s", conn.ReaderUnixSock.Path)
    return nil
}

// Add to your router
router := &server.Router{
    Groups: []server.GroupConfig{
        {
            Path: "/",
            Routes: []server.RouteConfig{
                {
                    Method:  http.MethodGet,
                    Path:    "/unixsock",
                    Handler: wsunixbridge.NewUpgradeHandler(socketsDir, connectionHandler),
                },
            },
        },
    },
}
```

**Initialization Event:**

When a WebSocket connection is established, the server automatically sends an initialization event with socket paths:

```go
// Client receives this event first:
{
    "type": "wsunixbridge.init",
    "data": {
        "writerSocket": "./sockets/f744bda5-1346-43a4-809b-6332e43fb993_output",
        "readerSocket": "./sockets/f744bda5-1346-43a4-809b-6332e43fb993_input"
    }
}
```

**External Tool Integration:**

Once you receive the initialization event, you can connect external tools to the Unix sockets:

```bash
# Read WebSocket data from external tools:
nc -U ./sockets/f744bda5-1346-43a4-809b-6332e43fb993_output
socat - UNIX-CONNECT:./sockets/f744bda5-1346-43a4-809b-6332e43fb993_output

# Send data to WebSocket from external tools:
echo "Hello from terminal!" | nc -U ./sockets/f744bda5-1346-43a4-809b-6332e43fb993_input
cat audio.mp3 | socat - UNIX-CONNECT:./sockets/f744bda5-1346-43a4-809b-6332e43fb993_input
```

**Use Cases:**

- Stream audio/video from external tools to WebSocket clients
- Send terminal output to web browsers in real-time
- Bridge legacy tools with modern web applications
- Real-time data processing pipelines
- IoT device integration

## Static Files & Uploads

**Static file serving using StaticRouteConfig:**

```go
Static: []server.StaticRouteConfig{
    {
        Dir:  "./public",
        Path: "/assets",
        DirectoryIndexingType: server.DirectoryIndexingTypeHTML, // or JSON, or None
    },
}
```

**File uploads with options:**

```go
Handler: s.FileUploadHandler("./uploads",
    server.WithFilenamePrependType(server.FilenamePrependTypeDateTime), // datetime_originalname.ext
    server.WithFileUploadHandlerPostprocessor(func(data map[string]any) (map[string]any, error) {
        // Process uploaded file data
        return data, nil
    }),
)
```

## Middleware System

**Built-in middleware:**

```go
GlobalMiddlewares: []middleware.Middleware{
    middleware.Recovery(),                    // Panic recovery
    middleware.RequestID(),                   // Request ID generation
    middleware.Logger(),                      // Request logging
    middleware.SecurityHeaders(),             // Security headers (XSS, CSRF, etc.)
    middleware.CORS(),                        // CORS with sensible defaults
    middleware.Timeout(30 * time.Second),     // Request timeout
    middleware.BasicAuth(map[string]string{"user": "pass"}), // Basic auth
}
```

**Per-group middleware using GroupConfig:**

```go
Groups: []server.GroupConfig{
    {
        Path: "/admin",
        Middlewares: []middleware.Middleware{
            middleware.BasicAuth(adminUsers),
        },
        Routes: []server.RouteConfig{...},
    },
}
```

## Built-in Handlers

**Health check:**

```go
{
    Method:  http.MethodGet,
    Path:    "/health",
    Handler: s.HealthHandler, // Returns {"status": "ok"}
}
```

**Echo endpoint:**

```go
{
    Method:  http.MethodPost,
    Path:    "/echo",
    Handler: s.EchoHandler, // Echoes request body back
}
```

## Complex Router Example

```go
router := &server.Router{
    // Global middleware applies to everything
    GlobalMiddlewares: []middleware.Middleware{
        middleware.Recovery(),
        middleware.RequestID(),
        middleware.Logger(),
        middleware.CORS(),
    },

    // Multiple static file routes
    Static: []server.StaticRouteConfig{
        {
            Dir:  "./public",
            Path: "/assets",
            DirectoryIndexingType: server.DirectoryIndexingTypeHTML,
        },
        {
            Dir:  "./uploads",
            Path: "/files",
            DirectoryIndexingType: server.DirectoryIndexingTypeJSON,
        },
    },

    Groups: []server.GroupConfig{
        // Public routes (no auth)
        {
            Path: "/",
            Routes: []server.RouteConfig{
                {Method: http.MethodGet, Path: "/health", Handler: healthHandler},
                {Method: http.MethodGet, Path: "/ws", Handler: dabluveees.UpgradeHandler(hub)},
            },
        },

        // API with JSON enforcement
        {
            Path: "/api/v1",
            Middlewares: []middleware.Middleware{
                middleware.EnforceRequestContentType("application/json"),
            },
            Routes: []server.RouteConfig{
                {Method: http.MethodGet, Path: "/users", Handler: getUsersHandler},
                {Method: http.MethodPost, Path: "/users", Handler: createUserHandler},
            },
        },

        // Admin routes with auth
        {
            Path: "/admin",
            Middlewares: []middleware.Middleware{
                middleware.BasicAuth(map[string]string{"admin": "secret"}),
            },
            Routes: []server.RouteConfig{
                {Method: http.MethodGet, Path: "/stats", Handler: adminStatsHandler},
                {Method: http.MethodDelete, Path: "/users/{id}", Handler: deleteUserHandler},
            },

            // Nested group for super admin
            Groups: []server.GroupConfig{
                {
                    Path: "/super",
                    Middlewares: []middleware.Middleware{
                        superAdminAuthMiddleware,
                    },
                    Routes: []server.RouteConfig{
                        {Method: http.MethodPost, Path: "/reset", Handler: systemResetHandler},
                    },
                },
            },
        },
    },
}
```

## Configuration

Environment variables (with sensible defaults):

```bash
export HTTP_SERVER_LISTENADDRESS="0.0.0.0:8080"         # HTTP server address
export HTTP_SERVER_TLSENABLED="true"                    # Enable TLS/HTTPS
export HTTP_SERVER_TLSLISTENADDRESS="0.0.0.0:8443"     # HTTPS server address
export HTTP_SERVER_TLSCERTFILE="/path/to/cert.pem"     # TLS certificate file
export HTTP_SERVER_TLSKEYFILE="/path/to/key.pem"       # TLS private key file
export HTTP_SERVER_READTIMEOUT="30s"                   # Request read timeout
export HTTP_SERVER_WRITETIMEOUT="30s"                  # Response write timeout
export HTTP_SERVER_IDLETIMEOUT="60s"                   # Connection idle timeout
export HTTP_SERVER_FILEUPLOADMAXMEMORY="33554432"      # Max upload memory in bytes (32MB)
```

Or use custom config:

```go
s, err := server.NewWithConfig(server.Config{
    ListenAddress: "127.0.0.1:9000",
    ReadTimeout:   10 * time.Second,
    WriteTimeout:  10 * time.Second,
})
```

## Troubleshooting

### Common Issues

**🚨 Server won't start / Address already in use**

```bash
# Error: bind: address already in use
# Solution: Use different port or kill the process using it
export HTTP_SERVER_LISTENADDRESS="127.0.0.1:8081"  # Different port
# or
sudo lsof -i :8080  # Find process using port 8080
kill -9 <PID>       # Kill the process
```

**🚨 WebSocket connections fail / CORS issues**

```go
// Make sure CORS is configured for WebSocket origins
GlobalMiddlewares: []middleware.Middleware{
    middleware.CORS(middleware.WithCORSAllowOrigins([]string{"https://mydomain.com"})),
}
```

**🚨 File uploads fail / Request body too large**

```bash
# Increase file upload memory limit (default is 32MB)
export HTTP_SERVER_FILEUPLOADMAXMEMORY="104857600"  # 100MB in bytes
```

**🚨 TLS/HTTPS server won't start**

```bash
# Make sure cert and key files exist and are readable
ls -la /path/to/cert.pem /path/to/key.pem
chmod 644 /path/to/cert.pem /path/to/key.pem

# Test with self-signed cert for development
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

**🚨 High memory usage / Memory leaks**

```go
// Make sure to properly close WebSocket hubs
defer hub.Close()

// Set reasonable timeouts
s, _ := server.NewWithConfig(server.Config{
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
})
```

**🚨 Static files not serving / 404 errors**

```go
// Make sure directory exists and is readable
Static: []server.StaticRouteConfig{{
    Dir:  "./public",  // Must exist
    Path: "/assets",   // URL prefix
}},
```

### Debug Tips

**Enable Debug Logging:**

```go
import "github.com/sirupsen/logrus"

logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)  // Enable debug logs
s, _ := server.NewWithLogger(logger)
```

**Check Server Status:**

```bash
# Health check endpoint (if enabled)
curl http://localhost:8080/health

# Check what's listening on your ports
netstat -tuln | grep :8080
```

**WebSocket Connection Testing:**

```javascript
// Browser console test for WebSocket connections
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onopen = () => console.log("Connected");
ws.onmessage = (e) => console.log("Message:", e.data);
ws.send(JSON.stringify({ type: "echo.request", data: "test" }));
```

## Production Deployment

### Docker Setup

**Dockerfile:**

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
EXPOSE 8080 8443
CMD ["./main"]
```

**docker-compose.yml:**

```yaml
version: "3.8"
services:
  app:
    build: .
    ports:
      - "8080:8080"
      - "8443:8443"
    environment:
      - HTTP_SERVER_LISTENADDRESS=0.0.0.0:8080
      - HTTP_SERVER_TLSENABLED=true
      - HTTP_SERVER_TLSLISTENADDRESS=0.0.0.0:8443
      - HTTP_SERVER_TLSCERTFILE=/certs/cert.pem
      - HTTP_SERVER_TLSKEYFILE=/certs/key.pem
    volumes:
      - ./certs:/certs:ro
      - ./uploads:/app/uploads
```

### Production Configuration

**Environment Variables:**

```bash
# Server binding (use 0.0.0.0 in containers)
export HTTP_SERVER_LISTENADDRESS="0.0.0.0:8080"

# Enable TLS/HTTPS in production
export HTTP_SERVER_TLSENABLED="true"
export HTTP_SERVER_TLSLISTENADDRESS="0.0.0.0:8443"
export HTTP_SERVER_TLSCERTFILE="/etc/ssl/certs/server.pem"
export HTTP_SERVER_TLSKEYFILE="/etc/ssl/private/server.key"

# Timeouts for production
export HTTP_SERVER_READTIMEOUT="30s"
export HTTP_SERVER_WRITETIMEOUT="30s"
export HTTP_SERVER_IDLETIMEOUT="120s"

# File upload limits
export HTTP_SERVER_FILEUPLOADMAXMEMORY="104857600"  # 100MB

# Service name for logging
export HTTP_SERVER_SERVICENAME="my-production-api"
```

**🚨 Security Warning - Built-in Handlers Are NOT Secured:**

The library includes security middleware but **built-in handlers have NO authentication by default**:

```go
// ⚠️  THESE ARE UNSECURED BY DEFAULT:
s.HealthHandler     // Anyone can access /health
s.EchoHandler       // Anyone can echo requests (exposes headers!)
s.FileUploadHandler // Anyone can upload files!

// ⚠️  WEBSOCKET ACCEPTS ALL ORIGINS BY DEFAULT:
websocket.UpgradeHandler(hub) // Allows connections from ANY website! (CSRF risk)
```

**You MUST add authentication to sensitive endpoints:**

```go
// ✅ SECURE VERSION - Add auth middleware to sensitive routes
Groups: []server.GroupConfig{
    {
        Path: "/",
        Routes: []server.RouteConfig{
            {Method: http.MethodGet, Path: "/health", Handler: s.HealthHandler}, // Public OK
        },
    },
    {
        Path: "/admin",
        Middlewares: []middleware.Middleware{
            middleware.BasicAuth(map[string]string{"admin": "secret123"}), // ADD AUTH!
        },
        Routes: []server.RouteConfig{
            {Method: http.MethodPost, Path: "/echo", Handler: s.EchoHandler},     // Now secured
            {Method: http.MethodPost, Path: "/upload", Handler: s.FileUploadHandler("./uploads")}, // Now secured
        },
    },
}

// ✅ SECURE WEBSOCKET - Configure CheckOrigin for production:
hub := wshub.NewHub("secure-hub")
secureUpgradeHandler := dabluveees.UpgradeHandler(hub, dabluveees.WithCheckOrigin(func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    // Only allow your trusted domains
    allowedOrigins := []string{
        "https://yourdomain.com",
        "https://app.yourdomain.com",
    }
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }
    return false // Reject all other origins
}))
```

**Security Best Practices:**

```go
// Production middleware stack
GlobalMiddlewares: []middleware.Middleware{
    middleware.Recovery(),                    // Panic recovery
    middleware.RequestID(),                   // Request tracing
    middleware.Logger(),                      // Request logging
    middleware.SecurityHeaders(),             // Security headers
    middleware.CORS(middleware.WithCORSAllowOrigins([]string{
        "https://yourdomain.com",             // Only allow your domain
    })),
    middleware.Timeout(30 * time.Second),     // Request timeout
},
```

**File Upload Security:**

- File uploads have **no size limits** by default except `HTTP_SERVER_FILEUPLOADMAXMEMORY`
- **No file type validation** - users can upload executables, scripts, etc.
- **No authentication** - anyone can upload if endpoint is exposed
- Files are stored with UUID prefixes to prevent overwrites, but **directory is world-readable**

```go
// Add your own validation:
Handler: s.FileUploadHandler("./uploads",
    server.WithFileUploadHandlerPostprocessor(func(data map[string]any) (map[string]any, error) {
        // Add your validation here:
        filename := data["filename"].(string)
        if !isAllowedFileType(filename) {
            return nil, fmt.Errorf("file type not allowed")
        }
        return data, nil
    }),
)
```

**WebSocket Security:**

- **CheckOrigin returns `true` for ALL origins by default** - allows any website to connect
- This creates **CSRF vulnerabilities** where malicious sites can connect to your WebSocket
- **No authentication** on WebSocket upgrade by default
- **All event handlers run without authentication** unless you add it manually

```go
// ⚠️  DEFAULT BEHAVIOR - DANGEROUS:
dabluveees.UpgradeHandler(hub) // Accepts connections from evil-site.com!

// ✅ PRODUCTION CONFIGURATION:
secureHandler := dabluveees.UpgradeHandler(hub,
    dabluveees.WithCheckOrigin(func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "https://yourtrustedsite.com"
    }),
)

// ✅ ADD AUTHENTICATION TO EVENT HANDLERS:
hub.RegisterEventHandler("sensitive.action", func(hub wshub.Hub, client *wshub.Client, event *dabluveees.Event) error {
    // Validate user permissions here before processing
    userID := client.GetUserID() // You need to implement this
    if !isAuthorized(userID, "sensitive.action") {
        return fmt.Errorf("unauthorized")
    }
    // Process event...
    return nil
})
```

**Graceful Shutdown:**

```go
func main() {
    s, err := server.New()
    if err != nil {
        log.Fatal(err)
    }

    // Setup signal handling for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan

        log.Println("Shutting down server...")
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer shutdownCancel()

        if err := s.Stop(shutdownCtx); err != nil {
            log.Printf("Error during shutdown: %v", err)
        }
        cancel()
    }()

    if err := s.Start(ctx, router); err != nil {
        log.Fatal(err)
    }
}
```

### Performance Tips

**Load Balancer Setup (nginx):**

```nginx
upstream backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;  # Multiple instances
}

server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

**Health Checks & Monitoring:**

```go
// Add health check endpoint
{
    Method:  http.MethodGet,
    Path:    "/health",
    Handler: s.HealthHandler,  // Returns {"status": "ok"}
}

// Add metrics endpoint (if using prometheus)
{
    Method:  http.MethodGet,
    Path:    "/metrics",
    Handler: promhttp.Handler(),
}
```

## Real Talk

This isn't another "minimal framework" that makes you implement everything. This is a **batteries-included HTTP utilities library** that handles the 90% of shit you always end up building anyway:

- 🔥 **Production-ready defaults** - TLS support, security headers, CORS, logging, graceful shutdown
- 🚀 **Zero-configuration startup** - Just create server, define routes, start
- 🛡️ **Security by default** - XSS protection, CSRF headers, secure defaults
- 📊 **Structured logging** - Request IDs, client IPs, timing, status codes
- 🌐 **CORS that works** - Sensible defaults, fully configurable
- 📁 **Static files + uploads** - Directory indexing, file caching, UUID filenames
- ⚡ **WebSocket support** - Hub system, event handling, broadcasting
- 🔧 **Completely customizable** - Override any default, add custom middleware
- 🧪 **90%+ test coverage** - Battle-tested and production-ready

## Why "aichteeteapee"?

Because saying "HTTP" is boring, but `aichteeteapee` makes you go "what the fuck is this?" and then you realize it's phonetically "HTTP" and you either laugh or hate it. Either way, you remember it.

Also, all the good names were taken.

## License

MIT - Because sharing is caring, and lawyers are expensive.

---

_"Finally, an HTTP library that doesn't make me want to switch careers."_ - Some Developer, Probably

_"I went from 200 lines of boilerplate to 20 lines of actual code."_ - Another Developer, Definitely

_"It just fucking works."_ - Everyone Who Uses This
