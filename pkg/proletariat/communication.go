package proletariat

import "time"

// Peer address
type Address string

// Basic configuration for the Communication instance.
// This will provide the parameters for binding the connection,
// timeout when handling messages.
type CommunicationConfiguration struct {
	// Address to bind the connection.
	Address Address

	// Timeout used when handling messages.
	Timeout time.Duration
}

// Base communication interface that should be implemented.
// This will be the interface the client will interact with,
// using the defined method is possible to send messages and
// to listen incoming messages.
type Communication interface {
	// Send the given data to the connect at the given address.
	Send(Address, []byte) error

	// Listen for incoming messages.
	Receive() <-chan []byte
}
