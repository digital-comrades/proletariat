package internal

import (
	"context"
	"github.com/jabolina/proletariat/pkg/proletariat"
	"net"
	"sync"
	"time"
)

const (
	minPollDelay = 5 * time.Millisecond
	maxPollDelay = 500 * time.Millisecond
)

// Default struct that implements the Communication interface.
// Using this implementation is possible to send and receive messages.
type DefaultCommunication struct {
	// Synchronize operations on available connections.
	mutex *sync.Mutex

	// Handler to carefully invoke new goroutines.
	handler *GoRoutineHandler

	// Configuration for the communication primitive.
	configuration proletariat.CommunicationConfiguration

	// Transport used to send and receive messages.
	transport Transport

	// Channel that will receive data from another connections.
	listener chan proletariat.Datagram

	// All established connections.
	connections map[proletariat.Address]Connection

	// Primitive context.
	ctx context.Context

	// Function to cancel the primitive execution.
	cancel context.CancelFunc
}

func NewCommunication(configuration proletariat.CommunicationConfiguration) (proletariat.Communication, error) {
	ctx, cancel := context.WithCancel(configuration.Ctx)
	tcp, err := NewTCPTransport(ctx, configuration.Address)
	if err != nil {
		cancel()
		return nil, err
	}

	comm := &DefaultCommunication{
		mutex:         &sync.Mutex{},
		handler:       NewRoutineHandler(),
		configuration: configuration,
		transport:     tcp,
		listener:      make(chan proletariat.Datagram),
		connections:   make(map[proletariat.Address]Connection),
		ctx:           ctx,
		cancel:        cancel,
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
		defer cancel()
		incoming := ConnectionConfiguration{
			Timeout:    d.configuration.Timeout,
			Read:       d.listener,
			Ctx:        ctx,
			Connection: conn,
			Target:     proletariat.Address(conn.RemoteAddr().String()),
		}
		connection := NewNetworkConnection(incoming)
		d.handler.Spawn(connection.Listen)
		<-ctx.Done()
	}
}

// For a given address, create a new connection instance if possible.
func (d *DefaultCommunication) resolveConnection(address proletariat.Address) (Connection, error) {
	if connection := d.getActiveConnection(address); connection != nil {
		return connection, nil
	}
	return d.establishNewConnection(address)
}

// Retrieve a connection for the in-memory available connections.
func (d *DefaultCommunication) getActiveConnection(address proletariat.Address) Connection {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.connections[address]
}

// Establish a connection with another peer using the available transport if possible.
func (d *DefaultCommunication) establishNewConnection(address proletariat.Address) (Connection, error) {
	conn, err := d.transport.Dial(address, d.configuration.Timeout)
	if err != nil {
		return nil, err
	}
	return d.saveNewConnection(conn), nil
}

// Given a connection, create a new proletariat.Connection and store it on the memory map.
func (d *DefaultCommunication) saveNewConnection(conn net.Conn) Connection {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	address := proletariat.Address(conn.RemoteAddr().String())
	config := ConnectionConfiguration{
		Timeout:    d.configuration.Timeout,
		Read:       d.listener,
		Connection: conn,
		Target:     address,
		Ctx:        d.ctx,
	}
	connection := NewNetworkConnection(config)
	d.connections[address] = connection
	return connection
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

// Accept new connections from external peers and start a new goroutine
// to start the life-cycle asynchronously.
// The Accept method to receive a new connection is a blocking call.
func (d *DefaultCommunication) poll() {
	var pollDelay = minPollDelay
	for {
		pollDelay = min(pollDelay*2, maxPollDelay)
		conn, err := d.transport.Accept()
		if err == nil {
			pollDelay = minPollDelay
			d.acceptIncomingConnection(conn)
		}

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(pollDelay):
			continue
		}
	}
}

// Implements the Communication interface.
func (d *DefaultCommunication) Close() error {
	defer d.handler.Close()
	d.cancel()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for _, connection := range d.connections {
		if err := connection.Close(); err != nil {
			return err
		}
	}
	return d.transport.Close()
}

// Implements the Communication interface.
func (d *DefaultCommunication) Start() {
	// Are we leaking?
	go d.poll()
}

// Implements the Communication interface.
func (d *DefaultCommunication) Send(address proletariat.Address, data []byte) error {
	connection, err := d.resolveConnection(address)
	if err != nil {
		return err
	}
	return connection.Write(data)
}

// Implements the Communication interface.
func (d *DefaultCommunication) Receive() <-chan proletariat.Datagram {
	return d.listener
}

func min(a time.Duration, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
