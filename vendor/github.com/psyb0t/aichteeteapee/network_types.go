package aichteeteapee

const (
	// Network types for net.Listen and similar functions.
	NetworkTypeTCP        = "tcp"        // TCP network type
	NetworkTypeTCP4       = "tcp4"       // TCP over IPv4
	NetworkTypeTCP6       = "tcp6"       // TCP over IPv6
	NetworkTypeUDP        = "udp"        // UDP network type
	NetworkTypeUDP4       = "udp4"       // UDP over IPv4
	NetworkTypeUDP6       = "udp6"       // UDP over IPv6
	NetworkTypeUnix       = "unix"       // Unix domain socket
	NetworkTypeUnixgram   = "unixgram"   // Unix datagram socket
	NetworkTypeUnixpacket = "unixpacket" // Unix packet socket
)
