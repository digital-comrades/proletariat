package internal

import (
	"github.com/jabolina/proletariat/pkg/proletariat"
	"net"
	"time"
)

// Create a UDP Transport.
// This struct will bind the server to the given address and will
// delegate commands to the listener.
type UDP struct {
	addr     net.Addr
	listener *net.UDPConn
}

// Create a new UDP Transport that will bind to the given address.
func NewUDPTransport(address *net.UDPAddr) (proletariat.Transport, error) {
	listening, err := net.ListenUDP("udp", address)
	if err != nil {
		return nil, err
	}
	udp := &UDP{
		listener: listening,
		addr:     address,
	}
	addr, ok := udp.Addr().(*net.UDPAddr)
	if !ok {
		defer listening.Close()
		return nil, proletariat.ErrNotUDP
	}

	if addr.IP == nil || addr.IP.IsUnspecified() {
		defer listening.Close()
		return nil, proletariat.ErrInvalidAddr
	}

	udp.addr = addr
	return udp, nil
}

// Implement Transport interface.
func (u *UDP) Accept() (net.Conn, error) {
	return u.listener, nil
}

// Implement Transport interface.
func (u *UDP) Close() error {
	return u.listener.Close()
}

// Implement Transport interface.
func (u *UDP) Addr() net.Addr {
	if u.addr != nil {
		return u.addr
	}
	return u.listener.LocalAddr()
}

// Implement Transport interface.
func (u *UDP) Bind(address proletariat.Address, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("udp", string(address), timeout)
}
