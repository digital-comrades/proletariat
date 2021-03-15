package proletariat

import "io"

// This interface represents a connection between two peers.
// The connection can be incoming, outgoing or duplex.
type Connection interface {
	io.Closer

	// Sends the encoded date to the target peer.
	Write([]byte) error

	// Start listening for incoming data.
	Listen()
}
