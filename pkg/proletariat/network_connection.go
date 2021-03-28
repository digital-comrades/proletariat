// Copyright (C) 2020-2021 digital-comrades and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proletariat

import (
	"bufio"
	"bytes"
	"context"
	"github.com/ugorji/go/codec"
	"net"
	"time"
)

const (
	ClosedConnection = "use of closed network connection"
)

// Gather all needed configuration for managing the connection.
type ConnectionConfiguration struct {
	// Timeout to apply for reading/writing to the connection.
	// Will only be applied if the value is greater than zero.
	Timeout time.Duration

	// Channel to publish the bytes received by the connection.
	Read chan<- Datagram

	// Parent context to bound the connection methods.
	Ctx context.Context

	// Cancel the current context.
	Cancel context.CancelFunc

	// The actual connection.
	Connection net.Conn

	// Peer address of the connection.
	Target Address
}

// Default Connection implementation.
// This will connect the peer to a target over the network.
// Commands will be sent/received using the available Conn.
type NetworkConnection struct {
	// Target peer.
	target Address

	// Established connection with the target.
	connection net.Conn

	// Reader to receive data from the connection.
	reader *bufio.Reader

	// Writer to send data to the connection.
	writer *bufio.Writer

	// Encodes all transported data.
	encoder *codec.Encoder

	// Decode received data.
	decoder *codec.Decoder

	// The configuration for the structure.
	configuration ConnectionConfiguration

	// Receiver is responsible for handling received messages.
	receiver IReceiver
}

func NewNetworkConnection(handler *GoRoutineHandler, configuration ConnectionConfiguration) Connection {
	r, w := bufio.NewReader(configuration.Connection), bufio.NewWriter(configuration.Connection)
	receiver := NewReceiver(configuration.Read, configuration.Timeout, configuration.Ctx)
	handler.Spawn(receiver.Start)
	return &NetworkConnection{
		configuration: configuration,
		target:        configuration.Target,
		connection:    configuration.Connection,
		reader:        r,
		writer:        w,
		encoder:       codec.NewEncoder(w, &codec.MsgpackHandle{}),
		decoder:       codec.NewDecoder(r, &codec.MsgpackHandle{}),
		receiver:      receiver,
	}
}

// Read data from the reader. The default buffer will have size 1 Kb.
// With this size, is possible that a message can be split in more than
// one buffer. The client must be careful when parsing the received bytes.
func (n *NetworkConnection) digest() ([]byte, error) {
	if n.configuration.Timeout > 0 {
		if err := n.connection.SetReadDeadline(time.Now().Add(n.configuration.Timeout)); err != nil {
			return nil, err
		}
	}

	// Decoding into a nil interface will delegate to the
	// decoder to correctly identify and construct the object.
	// In our case, a byte slice.
	var data interface{}
	if err := n.decoder.Decode(&data); err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	return data.([]byte), nil
}

// Implements the Connection interface.
func (n *NetworkConnection) Close() error {
	n.configuration.Cancel()
	if err := n.connection.Close(); err != nil {
		return err
	}
	return n.receiver.Close()
}

// Implements the Connection interface.
func (n *NetworkConnection) Write(bytes []byte) error {
	if n.configuration.Timeout > 0 {
		if err := n.connection.SetWriteDeadline(time.Now().Add(n.configuration.Timeout)); err != nil {
			return err
		}
	}

	if err := n.encoder.Encode(bytes); err != nil {
		return err
	}

	return n.writer.Flush()
}

// Implements the Connection interface.
// Digest bytes received from the underlining connection.
func (n *NetworkConnection) Listen() {
	for {
		select {
		case <-n.configuration.Ctx.Done():
			return
		default:
			data, err := n.digest()
			if data != nil {
				datagram := Datagram{
					Data: bytes.NewBuffer(data),
					Err:  err,
					From: Address(n.connection.RemoteAddr().String()),
					To:   Address(n.connection.LocalAddr().String()),
				}
				n.receiver.AddResponse(datagram)
			}
		}
	}
}
