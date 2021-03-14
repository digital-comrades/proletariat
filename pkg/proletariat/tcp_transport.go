package proletariat

import (
	"context"
	"net"
	"time"
)

// A TCP transport that implements the Transport interface.
// This struct will delegate the functions to the available listener.
type TCP struct {
	addr     net.Addr
	listener net.Listener
}

// Create a new TCP stream with the given address to bind.
func NewTCPTransport(parent context.Context, address Address) (Transport, error) {
	var lc net.ListenConfig
	listening, err := lc.Listen(parent, "tcp", string(address))
	if err != nil {
		return nil, err
	}
	tcp := &TCP{listener: listening}
	addr, ok := tcp.Addr().(*net.TCPAddr)
	if !ok {
		defer listening.Close()
		return nil, ErrNotTCP
	}

	if addr.IP == nil || addr.IP.IsUnspecified() {
		defer listening.Close()
		return nil, ErrInvalidAddr
	}
	tcp.addr = addr
	return tcp, nil
}

// Implement Transport interface.
func (t *TCP) Accept() (net.Conn, error) {
	return t.listener.Accept()
}

// Implement Transport interface.
func (t *TCP) Close() error {
	return t.listener.Close()
}

// Implement Transport interface.
func (t *TCP) Addr() net.Addr {
	if t.addr != nil {
		return t.addr
	}
	return t.listener.Addr()
}

// Implement Transport interface.
func (t *TCP) Dial(address Address, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", string(address), timeout)
}
