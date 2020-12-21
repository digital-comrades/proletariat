package proletariat

import (
	"errors"
	"net"
	"time"
)

var (
	ErrNotTCP      = errors.New("address is not TCP")
	ErrInvalidAddr = errors.New("address can not be used")

	ErrNotUDP = errors.New("address is not UDP")
)

type Transport interface {
	net.Listener

	Bind(address Address, timeout time.Duration) (net.Conn, error)
}
