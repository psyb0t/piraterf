package constants

const (
	// Client identifiers
	FieldClientID = "clientID"

	// WebSocket-specific fields
	FieldConnectionID = "connectionID"

	// Event-related fields
	FieldEventType = "eventType"
	FieldEventID   = "eventID"

	// Hub and system identifiers
	FieldHubName = "hubName"

	// Error and performance fields
	FieldTotalConns   = "totalConns"
	FieldTotalClients = "totalClients"
	FieldBufferSize   = "bufferSize"

	// Network and connection fields
	FieldRemoteAddr = "remoteAddr"
	FieldUserAgent  = "userAgent"
	FieldOrigin     = "origin"

	// WebSocket close fields
	FieldCloseCode = "closeCode"
	FieldCloseText = "closeText"

	// Configuration fields
	FieldReadBufferSize    = "readBufferSize"
	FieldWriteBufferSize   = "writeBufferSize"
	FieldHandshakeTimeout  = "handshakeTimeout"
	FieldEnableCompression = "enableCompression"
	FieldOldReadSize       = "oldReadSize"
	FieldOldWriteSize      = "oldWriteSize"
	FieldNewReadSize       = "newReadSize"
	FieldNewWriteSize      = "newWriteSize"

	// Server and endpoint fields
	FieldEndpoint = "endpoint"
)
