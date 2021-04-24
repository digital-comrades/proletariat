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
	"context"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	minPollDelay = 5 * time.Millisecond
	maxPollDelay = 500 * time.Millisecond
)

// DefaultCommunication default struct that implements the Communication interface.
// Using this implementation is possible to send and receive messages.
type DefaultCommunication struct {
	// Synchronize operations on available connections.
	mutex *sync.Mutex

	// Flag to transition between states.
	flag *Flag

	// Handler to carefully invoke new goroutines.
	handler *GoRoutineHandler

	// Configuration for the communication primitive.
	configuration Configuration

	// Transport used to send and receive messages.
	transport Transport

	// Channel that will receive data from another connections.
	listener chan Datagram

	// All established connections.
	connections map[Address][]Connection

	// Primitive context.
	ctx context.Context

	// Function to cancel the primitive execution.
	cancel context.CancelFunc

	// Channel to synchronize to primitive closing.
	closed chan bool
}

func NewCommunication(configuration Configuration) (Communication, error) {
	ctx, cancel := context.WithCancel(configuration.Ctx)
	tcp, err := NewTCPTransport(ctx, configuration.Address)
	if err != nil {
		cancel()
		return nil, err
	}

	comm := &DefaultCommunication{
		mutex:         &sync.Mutex{},
		flag:          &Flag{},
		handler:       NewRoutineHandler(),
		configuration: configuration,
		transport:     tcp,
		listener:      make(chan Datagram, 1024),
		connections:   make(map[Address][]Connection),
		ctx:           ctx,
		cancel:        cancel,
		closed:        make(chan bool, 1),
	}
	return comm, nil
}

// When a new connection request is received by the server this method is
// initiated. Using the given net connection a wrapper is created for this
// incoming request.
// This incoming connection request, will remain open until the peer closes,
// polling and for every received data will publish to the listener channel.
// After created, since this connection is created only to receive messages
// it will not be stored for the in-memory connections.
func (d *DefaultCommunication) handleIncomingConnection(conn net.Conn) {
	select {
	case <-d.ctx.Done():
		return
	default:
		ctx, cancel := context.WithCancel(d.ctx)
		incoming := ConnectionConfiguration{
			Timeout:    d.configuration.Timeout,
			Read:       d.listener,
			Ctx:        ctx,
			Cancel:     cancel,
			Connection: conn,
			Target:     Address(conn.RemoteAddr().String()),
		}
		connection := NewNetworkConnection(incoming)
		d.handler.Spawn(connection.Listen)
	}
}

// Verify if the communication is closed.
func (d *DefaultCommunication) isClosed() bool {
	select {
	case <-d.ctx.Done():
		return true
	default:
		return d.flag.IsInactive()
	}
}

// For a given address, create a new connection instance if possible.
func (d *DefaultCommunication) resolveConnection(address Address) (Connection, error) {
	if connection := d.getActiveConnection(address); connection != nil {
		return connection, nil
	}
	return d.establishNewConnection(address)
}

// Retrieve a connection for the in-memory available connections.
func (d *DefaultCommunication) getActiveConnection(address Address) Connection {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	connections, ok := d.connections[address]
	if !ok || len(connections) == 0 {
		return nil
	}

	var connection Connection
	size := len(connections)
	connection, connections[size-1] = connections[size-1], nil
	d.connections[address] = connections[:size-1]
	return connection
}

// Establish a connection with another peer using the available transport if possible.
func (d *DefaultCommunication) establishNewConnection(address Address) (Connection, error) {
	conn, err := d.transport.Dial(address, d.configuration.Timeout)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(d.ctx)
	config := ConnectionConfiguration{
		Timeout:    d.configuration.Timeout,
		Read:       d.listener,
		Connection: conn,
		Target:     address,
		Ctx:        ctx,
		Cancel:     cancel,
	}
	return NewNetworkConnection(config), nil
}

func (d *DefaultCommunication) maybeSaveConnection(address Address, connection Connection) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	available := d.connections[address]
	if d.configuration.PoolSize > 0 && len(available) > d.configuration.PoolSize {
		return
	}
	d.connections[address] = append(available, connection)
}

// Given a connection, create a new proletariat.Connection and store it on the memory map.
func (d *DefaultCommunication) saveNewConnection(conn net.Conn) {
	address := Address(conn.RemoteAddr().String())
	ctx, cancel := context.WithCancel(d.ctx)
	config := ConnectionConfiguration{
		Timeout:    d.configuration.Timeout,
		Read:       d.listener,
		Connection: conn,
		Target:     address,
		Ctx:        ctx,
		Cancel:     cancel,
	}
	d.maybeSaveConnection(address, NewNetworkConnection(config))
}

// Accept a incoming connection if the communication is not done.
func (d *DefaultCommunication) acceptIncomingConnection(conn net.Conn) {
	select {
	case <-d.ctx.Done():
		return
	default:
		d.saveNewConnection(conn)
		d.handler.Spawn(func() {
			d.handleIncomingConnection(conn)
		})
	}
}

// Close implements the Communication interface.
func (d *DefaultCommunication) Close() error {
	if d.flag.Inactivate() {
		defer d.handler.Close()
		d.cancel()
		d.mutex.Lock()
		defer d.mutex.Unlock()

		for key, connections := range d.connections {
			for _, connection := range connections {
				if err := connection.Close(); err != nil {
					return err
				}
			}
			delete(d.connections, key)
		}
		if err := d.transport.Close(); err != nil {
			return err
		}
		<-d.closed
		close(d.listener)
	}
	return nil
}

// Start implements the Communication interface.
// Accept new connections from external peers and start a new goroutine
// to start the life-cycle asynchronously.
// The Accept method to receive a new connection is a blocking call.
func (d *DefaultCommunication) Start() {
	if d.isClosed() {
		return
	}

	defer close(d.closed)
	var pollDelay = minPollDelay
	for {
		pollDelay = min(pollDelay*2, maxPollDelay)
		conn, err := d.transport.Accept()
		if err == nil {
			pollDelay = minPollDelay
			d.acceptIncomingConnection(conn)
		} else if strings.Contains(err.Error(), ClosedConnection) {
			d.cancel()
			d.closed <- true
		}

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(pollDelay):
			continue
		}
	}
}

// Send implements the Communication interface.
func (d *DefaultCommunication) Send(address Address, data []byte) error {
	if d.isClosed() {
		return ErrAlreadyClosed
	}

	connection, err := d.resolveConnection(address)
	if err != nil {
		return err
	}

	if err = connection.Write(data); err != nil {
		connection.Close()
		return err
	}
	d.maybeSaveConnection(address, connection)
	return nil
}

// Receive implements the Communication interface.
func (d *DefaultCommunication) Receive() <-chan Datagram {
	return d.listener
}

// Addr returns the current communication address.
func (d *DefaultCommunication) Addr() net.Addr {
	return d.transport.Addr()
}

func min(a time.Duration, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
