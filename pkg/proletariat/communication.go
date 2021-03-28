// Copyright (C) 2020-2021 digital-comrades and others
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
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"time"
)

var (
	ErrNotTCP        = errors.New("address is not TCP")
	ErrInvalidAddr   = errors.New("address can not be used")
	ErrAlreadyClosed = errors.New("communication was already closed")
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
	Data *bytes.Buffer

	// Errors received from the connection.
	Err error

	// Address that sent the message.
	From Address

	// Message destination.
	To Address
}
