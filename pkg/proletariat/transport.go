package proletariat

import (
	"net"
	"time"
)

// Basic transport that should be implemented.
type Transport interface {
	net.Listener

	// Dial to the given address to send requests.
	Dial(address Address, timeout time.Duration) (net.Conn, error)
}
