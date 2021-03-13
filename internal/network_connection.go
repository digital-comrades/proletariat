package internal

import (
	"bufio"
	"context"
	"github.com/jabolina/proletariat/pkg/proletariat"
	"net"
	"strings"
	"time"
)

const (
	closedConnection = "use of closed network connection"
)

// Gather all needed configuration for managing the connection.
type ConnectionConfiguration struct {
	// Timeout to apply for reading/writing to the connection.
	// Will only be applied if the value is greater than zero.
	Timeout time.Duration

	// Channel to publish the bytes received by the connection.
	Read chan<- proletariat.Datagram

	// Parent context to bound the connection methods.
	Ctx context.Context

	// The actual connection.
	Connection net.Conn

	// Peer address of the connection.
	Target proletariat.Address
}

// Default Connection implementation.
// This will connect the peer to a target over the network.
// Commands will be sent/received using the available Conn.
type NetworkConnection struct {
	// Target peer.
	target proletariat.Address

	// Established connection with the target.
	connection net.Conn

	// Reader to receive data from the connection.
	reader *bufio.Reader

	// Writer to send data to the connection.
	writer *bufio.Writer

	// The configuration for the structure.
	configuration ConnectionConfiguration
}

func NewNetworkConnection(configuration ConnectionConfiguration) Connection {
	return &NetworkConnection{
		configuration: configuration,
		target:        configuration.Target,
		connection:    configuration.Connection,
		reader:        bufio.NewReader(configuration.Connection),
		writer:        bufio.NewWriter(configuration.Connection),
	}
}

// Read data from the reader. The default buffer will have size 1 Kb.
// With this size, is possible that a message can be split in more than
// one buffer. The client must be careful when parsing the received bytes.
func (n *NetworkConnection) digest() ([]byte, error) {
	buffer := make([]byte, n.reader.Size())
	if n.configuration.Timeout > 0 {
		if err := n.connection.SetReadDeadline(time.Now().Add(n.configuration.Timeout)); err != nil {
			return nil, err
		}
	}

	size, err := n.reader.Read(buffer)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	}

	return buffer[:size], nil
}

// Implements the Connection interface.
func (n *NetworkConnection) Close() error {
	return n.connection.Close()
}

// Implements the Connection interface.
func (n *NetworkConnection) Write(bytes []byte) error {
	if n.configuration.Timeout > 0 {
		if err := n.connection.SetWriteDeadline(time.Now().Add(n.configuration.Timeout)); err != nil {
			return err
		}
	}

	if _, err := n.writer.Write(bytes); err != nil {
		return err
	}

	return n.writer.Flush()
}

// Implements the Connection interface.
// Digest bytes received from the underlining connection.
func (n *NetworkConnection) Listen() {
	defer n.connection.Close()

	for {
		select {
		case <-n.configuration.Ctx.Done():
			return
		default:
			data, err := n.digest()
			if err != nil {
				if strings.Contains(err.Error(), closedConnection) {
					return
				}
			}

			if data != nil {
				datagram := proletariat.Datagram{
					Data: data,
					Err:  err,
					From: proletariat.Address(n.connection.RemoteAddr().String()),
					To:   proletariat.Address(n.connection.LocalAddr().String()),
				}
				n.configuration.Read <- datagram
			}
		}
	}
}
