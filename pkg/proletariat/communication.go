package proletariat

import (
	"context"
	"errors"
	"io"
	"net"
	"time"
)

var (
	ErrNotTCP      = errors.New("address is not TCP")
	ErrInvalidAddr = errors.New("address can not be used")
)

// Peer address
type Address string

// Basic configuration for the Communication instance.
// This will provide the parameters for binding the connection,
// timeout when handling messages.
type Configuration struct {
	// Address to bind the connection.
	Address Address

	// Timeout used when handling messages.
	Timeout time.Duration

	// Connections can be pooled, this is the max size.
	PoolSize int

	// The parent context to handle the life-cycle of
	// the primitive.
	Ctx context.Context
}

// Base communication interface that should be implemented.
// This will be the interface the client will interact with,
// using the defined method is possible to send messages and
// to listen incoming messages.
type Communication interface {
	io.Closer

	// Start the Communication primitive, this method only return when
	// the context is closed and will run the whole life cycle.
	Start()

	// Send the given data to the connect at the given address.
	Send(Address, []byte) error

	// Listen for incoming messages.
	Receive() <-chan Datagram

	// Returns the current communication address.
	Addr() net.Addr
}

// Represent a datagram for the transport layer.
// Wraps the received data and errors from the connection.
type Datagram struct {
	// Received data from the underlining connection.
	Data []byte

	// Errors received from the connection.
	Err error

	// Address that sent the message.
	From Address

	// Message destination.
	To Address
}
