package internal

import (
	"github.com/jabolina/proletariat/pkg/proletariat"
	"net"
	"time"
)

// Basic transport that should be implemented.
type Transport interface {
	net.Listener

	// Dial to the given address to send requests.
	Dial(address proletariat.Address, timeout time.Duration) (net.Conn, error)
}
