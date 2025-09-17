package aichteeteapee

const (
	// Client identifiers
	FieldClientID = "clientID"

	// WebSocket-specific fields
	FieldConnectionID = "connectionID" // Individual connection identifier

	// Event-related fields
	FieldEventType = "eventType" // Event type being processed
	FieldEventID   = "eventID"   // Unique event identifier

	// Hub and system identifiers
	FieldHubName = "hubName" // Hub instance name

	// Error and performance fields
	FieldTotalConns   = "totalConns"   // Connection count context
	FieldTotalClients = "totalClients" // Connection count context
	FieldBufferSize   = "bufferSize"   // Buffer-related metrics

	// Network and connection fields
	FieldRemoteAddr = "remoteAddr" // Client remote address
	FieldUserAgent  = "userAgent"  // HTTP User-Agent header
	FieldOrigin     = "origin"     // WebSocket origin header

	// WebSocket close fields
	FieldCloseCode = "closeCode" // WebSocket close code
	FieldCloseText = "closeText" // WebSocket close text

	// Configuration fields
	FieldReadBufferSize    = "readBufferSize"    // WebSocket read buffer size
	FieldWriteBufferSize   = "writeBufferSize"   // WebSocket write buffer size
	FieldHandshakeTimeout  = "handshakeTimeout"  // WebSocket handshake timeout
	FieldEnableCompression = "enableCompression" // WebSocket compression setting
	FieldOldReadSize       = "oldReadSize"       // Previous read buffer size
	FieldOldWriteSize      = "oldWriteSize"      // Previous write buffer size
	FieldNewReadSize       = "newReadSize"       // New read buffer size
	FieldNewWriteSize      = "newWriteSize"      // New write buffer size

	// Server and endpoint fields
	FieldEndpoint = "endpoint" // HTTP endpoint path
)
